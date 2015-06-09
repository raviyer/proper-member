package node

import (
	"comms"
	"config"
	"fmt"
	"io"
	"log"
	"strconv"
	"sync"
	"time"
	"util"
)

const (config_version = 1
	STATUS_OK = 1
	STATUS_DOWN = -1)

type node struct {
	Guid      string
	Name      string
	Ip        string
	Port      int
	IsSeed    bool
	Seed      string
	Seedlings []string
	Peers     []string
	Version   int
	LastBeat  time.Time
	Status    int

	joinlock sync.WaitGroup
	fpmap    util.FailPointMap
}

func (this node) GetVersion() int {
	return this.Version
}
func (this *node) IncVersion() {
	this.Version += 1
}

func (this *node) updatePeer(guid string) {

	for _, m := range this.Peers {
		if m == guid {
			return
		}
	}
	this.Peers = append(this.Peers, guid)
}

func (this *node) updateSeedling(guid string) {
	for _, m := range this.Seedlings {
		if m == guid {
			return
		}
	}
	this.Seedlings = append(this.Seedlings, guid)
}

func updateNode(nodes *[]*node, n *node) {
	for i, v := range *nodes {
		if v.Guid == n.Guid {
			if v.GetVersion() <= n.GetVersion() {
				(*nodes)[i] = n
			}
			// The peer has not yet been updated. Just wait for
			// them to catch up
			return
		}
	}
	*nodes = append(*nodes, n)
}

type node_list struct {
	Nodes   []*node
	Version int
}

func (this node_list) GetVersion() int {
	return this.Version
}

func (this *node_list) IncVersion() {
	this.Version += 1
}

func (this *node_list) updateNode(n *node) {
	updateNode(&this.Nodes, n)
}

type global_config struct {
	Node    *node
	Members node_list
	Version int
}

func (this global_config) GetVersion() int {
	return this.Version
}

func (this *global_config) IncVersion() {
	this.Version += 1
}
func (this *global_config) updateNode(n *node) {
	if this.Node.Guid == n.Guid && this.Node.GetVersion() <= n.GetVersion() {
		this.Node = n
	}
	this.Members.updateNode(n)
}

func (this *global_config) getNodeByGuid(guid string) (n *node) {
	for _, m := range this.Members.Nodes {
		if m.Guid == guid {
			n = m
			break
		}
	}
	return
}

func (this *global_config) nameExists(name string) bool {
	for _, n := range this.Members.Nodes {
		if n.Name == name {
			return true
		}
	}
	return false
}

var gc *global_config

func (this *global_config) save() (err error) {
	return config.WriteConfig(configName(this.Node.Ip, this.Node.Port), this)
}
func configName(ip string, port int) string {
	return fmt.Sprintf("config_%v_%v.json", ip, port)
}

func (this *node) failpointName() string {
	return fmt.Sprintf("failpoints_%v_%v.json", this.Ip, this.Port)
}

func NewNode(ip string, port int) (nodeOut *node, err error) {
	// Read config check if it exists for node/port, initialize
	cf := configName(ip, port)
	gc = new(global_config)
	err = config.ReadConfig(cf, gc)
	var guid string
	if err != nil {
		gc.Version = config_version
		log.Printf("Not found config creating\n")
		tmpn := new(node)
		// If not create GUID and Init to defaults
		guid, err = util.NewUUID()
		if err != nil {
			return
		}
		tmpn = new(node)
		tmpn.Guid = guid
		tmpn.Name = ""
		tmpn.Ip = ip
		tmpn.Port = port
		tmpn.IsSeed = false
		tmpn.Seed = ""
		tmpn.Peers = *new([]string)
		tmpn.Seedlings = *new([]string)
		tmpn.LastBeat = time.Time{}
		tmpn.Status = STATUS_OK
		tmpn.fpmap = util.NewFailPointMap(tmpn.failpointName())
		tmpn.initFailpoints()
		err = tmpn.fpmap.Save(tmpn.failpointName())

		gc.Node = tmpn
		gc.updateNode(tmpn)
		err = gc.save()
		if err != nil {
			return
		}
	} else {
		gc.Node.fpmap = util.NewFailPointMap(gc.Node.failpointName())
		fmt.Printf("%+v\n", gc.Node.fpmap)

		log.Printf("found config %+v\n", gc.Node)
	}

	nodeOut = gc.Node
	return
}

type nodeURIHandler func(*node, io.Writer, comms.Request) error

func (n *node) members(resw io.Writer, req comms.Request) (err error) {

	return config.EncodeObject(resw, &gc.Members)
}

func (this *node) attach(resw io.Writer, req comms.Request) (err error) {
	this.joinlock.Add(1)
	defer this.joinlock.Done()
	// Get the remote's host and port from request
	host := req.GetParameter("host")
	port, err := strconv.Atoi(req.GetParameter("port"))
	if err != nil {
		return
	}
	log.Printf("Got attach for %v:%v\n", host, port)
	// Verify node exists
	tmpn := new(node)
	clnt := comms.NewClient(host, port)
	err = clnt.Get("nodeinfo", nil, tmpn)
	if err != nil {
		log.Printf("Failed to attach: %v:%v Error %v", host, port, err)
		return
	}
	log.Printf("Peers is %+v\n", *tmpn)
	tmpn.Seed = this.Guid
	// Update Seedlings and the Node List
	this.updateSeedling(tmpn.Guid)
 	// You are now a seed
	this.IsSeed = true
	// Name ourselves and the seed if required
	stringp := []*string{&this.Name, &tmpn.Name}
	for _, s := range stringp {
		if len(*s) == 0 {
			var name string
			for {
				name, err = util.GetRandomName()
				if err != nil {
					return
				}
				if !gc.nameExists(name) {
					break
				}
			}
			*s = name
		}
	}
	this.IncVersion()
	tmpn.IncVersion()
	gc.updateNode(this)
	gc.updateNode(tmpn)

	err = clnt.Set("join", nil, gc)
	if err != nil {
		return
	}

	// Save the config
	err = gc.save()

	return
}

func (this *node) join(resw io.Writer, req comms.Request) (err error) {
	this.joinlock.Add(1)
	defer this.joinlock.Done()
	peergc := new(global_config)
	defer req.GetBody().Close()
	err = config.DecodeObject(req.GetBody(), peergc)
	if err != nil {
		log.Printf("Error decoding node information %v", err)
		return
	}
	log.Printf("Joining up with %+v", *peergc.Node)
	this.Seed = peergc.Node.Guid
	for _, v := range peergc.Members.Nodes {
		gc.updateNode(v)
	}

	for _, m := range peergc.Node.Seedlings {
		this.updatePeer(m)
	}

	err = gc.save()
	if err != nil {
		log.Printf("Error decoding node information %v", err)
		return
	}
	return
}

func (n *node) fail(resw io.Writer, req comms.Request) (err error) {
	fmt.Fprintf(resw, "fail!\n")
	return
}

func (n *node) dump(resw io.Writer, req comms.Request) (err error) {
	log.Printf("Start - Dump Request")
	fmt.Fprintf(resw, "%+v\n", n)
	log.Printf("Done - Dump Request")
	return
}

func (n *node) nodeinfo(resw io.Writer, req comms.Request) (err error) {
	return config.EncodeObject(resw, n)
}

func (this *node) ping(resw io.Writer, req comms.Request) (err error) {
	this.joinlock.Add(1)
	defer this.joinlock.Done()

	this.LastBeat = time.Now()
	tmpgc := new(global_config)
	defer req.GetBody().Close()
	err = config.DecodeObject(req.GetBody(), tmpgc)
	if err != nil {
		log.Printf("Error decoding node information %v", err)
		return
	}

	if !this.IsSeed || len(this.Peers) > 0 {
		for _, v := range tmpgc.Node.Seedlings {
			this.updatePeer(v)
		}
	}
	for _, m := range tmpgc.Members.Nodes {
		gc.updateNode(m)
	}
	gc.save()

	return
}

func (n *node) ola(resw io.Writer, req comms.Request) (err error) {
	fmt.Fprintf(resw, "Ola!\n")
	return
}

type fpdata struct {
	Name        string
	Enabled     bool
	Probability int
	Err       string
	Version int
}

func (this fpdata) GetVersion() int {
	return 1
}
func (this fpdata) IncVersion() {
}

func (this *node) failpoint(resw io.Writer, req comms.Request) (err error) {
	method := req.GetMethod()
	switch {
	case method == "GET":
		config.EncodeObject(resw, &this.fpmap)
	case method == "POST":
		var data fpdata
		defer req.GetBody().Close()
		err = config.DecodeObject(req.GetBody(), &data)
		if err != nil {
			return
		}
		this.fpmap.UpdateFailpoint(data.Name, data.Enabled,
			data.Probability, data.Err)
		log.Printf("Trying to save to %s", this.failpointName())
		this.fpmap.Save(this.failpointName())
	}
	return
}

var nodeHandlerMap = map[string]nodeURIHandler{
	"members":   (*node).members,
	"ping":      (*node).ping,
	"attach":    (*node).attach,
	"join":      (*node).join,
	"fail":      (*node).fail,
	"dump":      (*node).dump,
	"nodeinfo":  (*node).nodeinfo,
	"failpoint": (*node).failpoint,
	"":          (*node).ola,
}

func nodeHandler(n *node, uri string) func(io.Writer, comms.Request) (err error) {
	handler := func(resw io.Writer, req comms.Request) (err error) {
		funk, ok := nodeHandlerMap[uri]
		if ok {
			return funk(n, resw, req)
		} else {
			return
		}
	}
	return handler
}

/*
 * You will send a heartbeat to up to 2 nodes, if you are only a Seed or Only
 * a Peer then you send to the next  Peer. If you are both a Seed and a Peer
 * to another Seed you will send a ping to both nodes.
 */
func (this *node) findBeat() (nodes *[]*node) {
	nodes = new([]*node)
	// If you are a seed pick up the first sibling to ping
	if this.IsSeed {
		if len(this.Seedlings) > 0 {
			*nodes = append(*nodes,
				gc.getNodeByGuid(this.Seedlings[0]))
		}
	}
	// If you have peers ping them or if you are the last peer ping
	// your Seed
	if len(this.Peers) > 0 {
		var i int
		var p string
		for i, p = range this.Peers {
			if p == this.Guid {
				break
			}
		}
		var n *node
		if i+1 == len(this.Peers) {
			n = gc.getNodeByGuid(this.Seed)
		} else {
			n = gc.getNodeByGuid(this.Peers[i+1])
		}
		*nodes = append(*nodes, n)
	}

	return
}

func (this *node) heartBeat() {
	// If you are the seed, then send to your first seedling,
	// If a seedling then to the Next peer
	for {
		var err error
		time.Sleep(time.Second)
		this.joinlock.Wait()
		nodes := this.findBeat()
		if len(*nodes) == 0 {
			continue
		}
		for _, p := range *nodes {
			comms.NewClient(p.Ip, p.Port).Set("ping", nil, gc)

			if err != nil {
				log.Printf("Error pinging %v:%v - %v",
					p.Ip, p.Port, err)
			}
		}
	}
}
func (this *node) initFailpoints() {
	fmt.Printf("%+v\n", this.fpmap)
	this.fpmap.AddFailpoint("FailToPing", util.Disabled, 100, "Failed to ping")
}

func (this *node) Start() (err error) {
	srv := comms.NewServer(this.Ip, this.Port)
	for key, _ := range nodeHandlerMap {
		srv.RegisterHandler(key, nodeHandler(this, key))
	}

	go this.heartBeat()
	return srv.Start()
}

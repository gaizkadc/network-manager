/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

 package entities

type Network struct {
	OrganizationId string
	NetworkId string
	Name string
	CreationTimestamp int64
}

type DNSEntry struct {
	NetworkId string
	FQDN string
	IP string
}

func NewNode(n v1.Node) *Node {
	ip := "NotFound"
	if len(n.Status.Addresses) > 0 {
		ip = n.Status.Addresses[0].Address
	}
	return &Node{
		IP:     ip,
		Labels: n.Labels,
	}
}
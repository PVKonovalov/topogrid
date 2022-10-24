# topogrid
Package topogrid contains implementations of basic power grid algorithms based on the grid topology.
We use three main things - node, edge and equipment. Each power equipment can be represented as a topological node or edge.

## Database
The power system topology is stored in the database as a set of tables.
![Configuration database schema](database/TopoGridDatabase.png)
## Using
```golang
topology := topogrid.New(len(nodes))

for _, node := range nodes {
  topology.AddNode(node.Id, 
    node.EquipmentId, 
    node.EquipmentTypeId, 
    node.EquipmentName)
}

for _, edge := range edges {
  err := topology.AddEdge(edge.Id, 
    edge.Terminal1, 
    edge.Terminal2, 
    edge.StateNormal, 
    edge.EquipmentId, 
    edge.EquipmentTypeId, 
    edge.EquipmentName)
  if err != nil {
    log.Errorf("%v", err)
  }
}
```
### NodeIsPoweredBy
Get an array of nodes id with the type of equipment "TypePower" from which the specified node is powered with the current electrical state (On/Off) of the circuit breakers
```golang
for _, node := range nodes {
  poweredBy, err := topology.NodeIsPoweredBy(node.Id)
    if err != nil {
      log.Errorf("%v", err)
    }
    log.Debugf("%d:%s <- %v:%s", node.Id, topology.EquipmentNameByNodeId(node.Id), poweredBy, topology.EquipmentNameByNodeIdArray(poweredBy))
}
```
### NodeCanBePoweredBy 
Get an array of nodes id with the type of equipment "Power", from which the specified node can be powered regardless of the current electrical state (On/Off) of the circuit breakers
```golang
for _, node := range nodes {
  poweredBy, err := topology.NodeCanBePoweredBy(node.Id)
    if err != nil {
      log.Errorf("%v", err)
    }
    log.Debugf("%d:%s <- %v:%s", node.Id, topology.EquipmentNameByNodeId(node.Id), poweredBy, topology.EquipmentNameByNodeIdArray(poweredBy))
}
```
### CircuitBreakersNextToNode 
Get an array of IDs of circuit breakers next to the node. If we need to isolate some area of the electrical network, we need to find all circuit breakers near a node in that area.
```golang
for _, node := range nodes {
  nextTo, err := topology.CircuitBreakersNextToNode(node.Id)
    if err != nil {
      log.Errorf("%v", err)
    }
    log.Debugf("%d:%s <- %v:%s", node.Id, topology.EquipmentNameByNodeId(node.Id), poweredBy, topology.EquipmentNameByNodeIdArray(nextTo))
}
```


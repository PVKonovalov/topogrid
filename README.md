# Topogrid
Package topogrid contains implementations of basic power grid algorithms based on the grid topology.
We use three main things - node, edge and equipment. Each power equipment can be represented as a topological node or edge.
The [wonderful library](https://github.com/yourbasic/graph) is used to represent the graph.

## List of terms and abbreviations
* Edge: A link between two nodes. From the point of view of electrical network equipment, edge can imagine circuit 
breakers, disconnectors, power transformers, earthing switches, etc. All electrical network equipment with more 
than one terminal.
* Node: The name for any single junction. Nodes are connected to one another by edges. E.g. power supply, DERs, 
Consumer transformer substations, ground (earth) etc. 
* Terminal: The endpoint of a power grid equipment, represented by node.
## Distribution grid example
![Configuration database schema](assets/ExampleGrid.png)

## Graph example
![Graph example](assets/ExampleGridGraph.svg)
## Database
The power system topology is stored in the database as a set of tables.
![Configuration database schema](assets/TopoGridDatabase.png)
## Usage

```go
type EquipmentStruct struct {
	id              int
	typeId          int
	name            string
	electricalState uint8
	poweredBy       map[int]int64
	switchState     int
}

type NodeStruct struct {
	idx             int
	id              int
	equipmentId     int
	electricalState uint8
}

type TerminalStruct struct {
	node1Id          int
	node2Id          int
	numberOfSwitches int64
}

type EdgeStruct struct {
	idx         int
	id          int
	equipmentId int
	terminal    TerminalStruct
}
```
```go
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
### EquipmentNameByEquipmentId
Returns a string with node name from the equipment id
```go
func (t *TopologyGridStruct) EquipmentNameByEquipmentId(equipmentId int) string
```

### EquipmentNameByNodeIdx 
Returns a string with node name from the node index
```go
func (t *TopologyGridStruct) EquipmentNameByNodeIdx(idx int) string
```

###  EquipmentNameByNodeId
Returns a string with node name from the node id
```go
func (t *TopologyGridStruct) EquipmentNameByNodeId(id int) string
```

### EquipmentNameByNodeIdArray
Returns a string with node names separated by ',' from an array of node ids
```go
func (t *TopologyGridStruct) EquipmentNameByNodeIdArray(idArray []int) string
```

### EquipmentNameByEdgeIdx
Returns a string with node name from the node index
```go
func (t *TopologyGridStruct) EquipmentNameByEdgeIdx(idx int) string
```

### EquipmentNameByEdgeId
Returns a string with node name from the node id
```go
func (t *TopologyGridStruct) EquipmentNameByEdgeId(id int) string
```

### EquipmentNameByEdgeIdArray
Returns a string with node names separated by ',' from an array of node ids
```go
func (t *TopologyGridStruct) EquipmentNameByEdgeIdArray(idArray []int) string
```

### EquipmentIdByEdgeId
Returns equipment identifier by corresponded edge id
```go
func (t *TopologyGridStruct) EquipmentIdByEdgeId(edgeId int) (int, error)
```

### SetSwitchStateByEquipmentId
Set switchState field and changes current topology graph
```go
func (t *TopologyGridStruct) SetSwitchStateByEquipmentId(equipmentId int, switchState int) error
```

### AddNode
Add node to grid topology
```go
func (t *TopologyGridStruct) AddNode(id int, equipmentId int, equipmentTypeId int, equipmentName string)
```

### AddEdge
Add edge to grid topology
```go
func (t *TopologyGridStruct) AddEdge(id int, terminal1 int, terminal2 int, state int, equipmentId int, equipmentTypeId int, equipmentName string) error
```

### NodeIsPoweredBy
Get an array of nodes id with the type of equipment "TypePower" from which the specified node is powered with the current 'switchState' (On/Off) of the circuit breakers
```go
func (t *TopologyGridStruct) NodeIsPoweredBy(nodeId int) ([]int, error)
```
```go
for _, node := range nodes {
  poweredBy, err := topology.NodeIsPoweredBy(node.Id)
    if err != nil {
      log.Errorf("%v", err)
    }
    log.Debugf("%d:%s <- %v:%s", node.Id, topology.EquipmentNameByNodeId(node.Id), poweredBy, topology.EquipmentNameByNodeIdArray(poweredBy))
}
```
### NodeCanBePoweredBy 
Get an array of nodes id with the type of equipment "Power", from which the specified node can be powered regardless of the current 'switchState'  (On/Off) of the circuit breakers
```go
func (t *TopologyGridStruct) NodeCanBePoweredBy(nodeId int) ([]int, error)
```
```go
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
![Next to node](assets/NextToNode.png)

```go
func (t *TopologyGridStruct) GetCircuitBreakersEdgeIdsNextToNode(nodeId int) ([]int, error)
```
```go
for _, node := range nodes {
  nextTo, err := topology.CircuitBreakersNextToNode(node.Id)
    if err != nil {
      log.Errorf("%v", err)
    }
    log.Debugf("%d:%s <- %v:%s", node.Id, topology.EquipmentNameByNodeId(node.Id), poweredBy, topology.EquipmentNameByNodeIdArray(nextTo))
}
```
### BfsFromNodeId 
Traverses current graph in breadth-first order starting at nodeStart
```go
func (t *TopologyGridStruct) BfsFromNodeId(nodeIdStart int) []TerminalStruct 
```
### GetAsGraphMl 
Returns a string with a graph represented by the [graph modeling language](https://en.wikipedia.org/wiki/Graph_Modelling_Language) 
```go
func (t *TopologyGridStruct) GetAsGraphMl() string 
```

### SetEquipmentElectricalState
Set electrical states for equipment. Use this method to set colors on your single line diagram (SLD).
![Configuration database schema](assets/ElectricalState.svg)
```go
// Equipment electrical states
const (
	StateIsolated    uint8 = 0x00
	StateEnergized   uint8 = 0x01
	StateGrounded    uint8 = 0x02
	StateOvercurrent uint8 = 0x04
	StateFault       uint8 = 0x08
)
```
```go
func (t *TopologyGridStruct) SetEquipmentElectricalState()
```




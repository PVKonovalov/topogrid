// Package topogrid contains implementations of basic power grid algorithms based on the grid topology.
//

package topogrid

import (
	"errors"
	"fmt"
	"grid_test/graph"
	"sync"
)

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

type TopologyGridStruct struct {
	sync.RWMutex

	currentGraph *graph.Mutable // Current grid topology (depends on circuit breaker states)
	fullGraph    *graph.Mutable // Full grid topology

	nodes     []NodeStruct
	edges     []EdgeStruct
	equipment map[int]EquipmentStruct

	nodeIdxFromNodeId              map[int]int   // NodeId -> NodeIdx
	nodeIdArrayFromEquipmentTypeId map[int][]int // EquipmentTypeId -> []NodeId
	nodeIdArrayFromEquipmentId     map[int][]int // EquipmentId -> []NodeId

	edgeIdxFromEdgeId              map[int]int              // EdgeId -> EdgeIdx
	edgeIdArrayFromEquipmentTypeId map[int][]int            // EquipmentTypeId -> []EdgeId
	edgeIdArrayFromTerminalStruct  map[TerminalStruct][]int // TerminalStruct -> []EdgeId
	edgeIdArrayFromNodeId          map[int][]int            // NodeId -> []EdgeId
	nodeIdx                        int
	edgeIdx                        int
}

// New topology
func New(numberOfNodes int) *TopologyGridStruct {
	return &TopologyGridStruct{
		currentGraph:                   graph.New(numberOfNodes),
		fullGraph:                      graph.New(numberOfNodes),
		nodes:                          make([]NodeStruct, numberOfNodes),
		nodeIdxFromNodeId:              make(map[int]int),
		nodeIdArrayFromEquipmentTypeId: make(map[int][]int),
		nodeIdArrayFromEquipmentId:     make(map[int][]int),
		edgeIdArrayFromEquipmentTypeId: make(map[int][]int),
		edgeIdxFromEdgeId:              make(map[int]int),
		edgeIdArrayFromTerminalStruct:  make(map[TerminalStruct][]int),
		edgeIdArrayFromNodeId:          make(map[int][]int),
		edges:                          make([]EdgeStruct, 0),
		nodeIdx:                        0,
		edgeIdx:                        0,
		equipment:                      make(map[int]EquipmentStruct),
	}
}

// EquipmentNameByEquipmentId returns a string with node name from the equipment id
func (t *TopologyGridStruct) EquipmentNameByEquipmentId(equipmentId int) string {
	return t.equipment[equipmentId].name
}

// EquipmentNameByNodeIdx returns a string with node name from the node index
func (t *TopologyGridStruct) EquipmentNameByNodeIdx(idx int) string {
	return t.equipment[t.nodes[idx].equipmentId].name
}

// EquipmentNameByNodeId returns a string with node name from the node id
func (t *TopologyGridStruct) EquipmentNameByNodeId(id int) string {
	if idx, exists := t.nodeIdxFromNodeId[id]; exists {
		return t.EquipmentNameByNodeIdx(idx)
	} else {
		return ""
	}
}

//EquipmentNameByNodeIdArray returns a string with node names separated by ',' from an array of node ids
func (t *TopologyGridStruct) EquipmentNameByNodeIdArray(idArray []int) string {
	var name string
	for i, id := range idArray {
		if i != 0 {
			name += ","
		}
		name += t.EquipmentNameByNodeId(id)
	}
	return name
}

// EquipmentNameByEdgeIdx returns a string with node name from the node index
func (t *TopologyGridStruct) EquipmentNameByEdgeIdx(idx int) string {
	return t.equipment[t.edges[idx].equipmentId].name
}

// EquipmentNameByEdgeId returns a string with node name from the node id
func (t *TopologyGridStruct) EquipmentNameByEdgeId(id int) string {
	if idx, exists := t.edgeIdxFromEdgeId[id]; exists {
		return t.EquipmentNameByEdgeIdx(idx)
	} else {
		return ""
	}
}

// EquipmentNameByEdgeIdArray returns a string with node names separated by ',' from an array of node ids
func (t *TopologyGridStruct) EquipmentNameByEdgeIdArray(idArray []int) string {
	var name string
	for i, id := range idArray {
		if i != 0 {
			name += ","
		}
		name += t.EquipmentNameByEdgeId(id)
	}
	return name
}

// AddNode to grid topology
func (t *TopologyGridStruct) AddNode(id int, equipmentId int, equipmentTypeId int, equipmentName string) {

	if equipmentId != 0 {
		t.equipment[equipmentId] = EquipmentStruct{
			id:              equipmentId,
			typeId:          equipmentTypeId,
			name:            equipmentName,
			electricalState: StateIsolated,
			poweredBy:       make(map[int]int64),
		}
	}

	t.nodes[t.nodeIdx] = NodeStruct{idx: t.nodeIdx, id: id, equipmentId: equipmentId}

	t.nodeIdxFromNodeId[id] = t.nodeIdx

	if _, exists := t.nodeIdArrayFromEquipmentId[equipmentId]; !exists {
		t.nodeIdArrayFromEquipmentId[equipmentId] = make([]int, 0)
	}
	t.nodeIdArrayFromEquipmentId[equipmentId] = append(t.nodeIdArrayFromEquipmentId[equipmentId], id)

	if _, exists := t.nodeIdArrayFromEquipmentTypeId[equipmentTypeId]; !exists {
		t.nodeIdArrayFromEquipmentTypeId[equipmentTypeId] = make([]int, 0)
	}
	t.nodeIdArrayFromEquipmentTypeId[equipmentTypeId] = append(t.nodeIdArrayFromEquipmentTypeId[equipmentTypeId], id)

	t.nodeIdx += 1
}

// AddEdge to grid topology
func (t *TopologyGridStruct) AddEdge(id int, terminal1 int, terminal2 int, state int, equipmentId int, equipmentTypeId int, equipmentName string) error {
	terminal := TerminalStruct{node1Id: terminal1, node2Id: terminal2}
	t.edges = append(t.edges,
		EdgeStruct{idx: t.edgeIdx,
			id:          id,
			equipmentId: equipmentId,
			terminal:    terminal,
		})

	if equipmentId != 0 {
		t.equipment[equipmentId] = EquipmentStruct{id: equipmentId,
			typeId:          equipmentTypeId,
			name:            equipmentName,
			electricalState: StateIsolated,
			poweredBy:       make(map[int]int64),
			switchState:     state,
		}
	}

	t.edgeIdxFromEdgeId[id] = t.edgeIdx

	if _, exists := t.nodeIdArrayFromEquipmentId[equipmentId]; !exists {
		t.nodeIdArrayFromEquipmentId[equipmentId] = make([]int, 0)
	}
	t.nodeIdArrayFromEquipmentId[equipmentId] = append(t.nodeIdArrayFromEquipmentId[equipmentId], terminal1)
	t.nodeIdArrayFromEquipmentId[equipmentId] = append(t.nodeIdArrayFromEquipmentId[equipmentId], terminal2)

	if _, exists := t.edgeIdArrayFromTerminalStruct[terminal]; !exists {
		t.edgeIdArrayFromTerminalStruct[terminal] = make([]int, 0)
	}

	t.edgeIdArrayFromTerminalStruct[terminal] = append(t.edgeIdArrayFromTerminalStruct[terminal], id)

	if _, exists := t.edgeIdArrayFromEquipmentTypeId[equipmentTypeId]; !exists {
		t.edgeIdArrayFromEquipmentTypeId[equipmentTypeId] = make([]int, 0)
	}

	t.edgeIdArrayFromEquipmentTypeId[equipmentTypeId] = append(t.edgeIdArrayFromEquipmentTypeId[equipmentTypeId], id)

	if _, exists := t.edgeIdArrayFromNodeId[terminal1]; !exists {
		t.edgeIdArrayFromNodeId[terminal1] = make([]int, 0)
	}

	t.edgeIdArrayFromNodeId[terminal1] = append(t.edgeIdArrayFromNodeId[terminal1], id)

	if _, exists := t.edgeIdArrayFromNodeId[terminal2]; !exists {
		t.edgeIdArrayFromNodeId[terminal2] = make([]int, 0)
	}

	t.edgeIdArrayFromNodeId[terminal2] = append(t.edgeIdArrayFromNodeId[terminal2], id)

	t.edgeIdx += 1

	node1idx, existsNode1 := t.nodeIdxFromNodeId[terminal1]
	node2idx, existsNode2 := t.nodeIdxFromNodeId[terminal2]

	// Edge cost == 0 but for Circuit Breaker cost == 1, so we can calculate the shortest path between two nodes
	// to know how many CBs between ones
	var cost int64 = 0
	if equipmentTypeId == TypeCircuitBreaker {
		cost = 1
	}

	if existsNode1 && existsNode2 {
		if state == 1 {
			t.currentGraph.AddBothCost(node1idx, node2idx, cost)
		}

		if equipmentTypeId != TypeDisconnectSwitch || (equipmentTypeId == TypeDisconnectSwitch && state == 1) {
			t.fullGraph.AddBothCost(node1idx, node2idx, cost)
		}

	} else {
		return errors.New(fmt.Sprintf("Nodes %d:%d are not found", terminal1, terminal2))
	}

	return nil
}

// NodeIsPoweredBy returns an array of nodes id with the type of equipment "TypePower"
// from which the specified node is powered with the current electricalState of the circuit breakers
func (t *TopologyGridStruct) NodeIsPoweredBy(nodeId int) ([]int, error) {
	poweredBy := make([]int, 0)

	nodeIdx, exists := t.nodeIdxFromNodeId[nodeId]

	if !exists {
		return nil, errors.New(fmt.Sprintf("node idx was not found for node id %d", nodeId))
	}

	for _, nodeTypePowerId := range t.nodeIdArrayFromEquipmentTypeId[TypePower] {

		nodeTypePowerIdx, exists := t.nodeIdxFromNodeId[nodeTypePowerId]

		if !exists {
			return nil, errors.New(fmt.Sprintf("node idx was not found for node id %d", nodeId))
		}

		path, _ := graph.ShortestPath(t.currentGraph, nodeTypePowerIdx, nodeIdx)
		if len(path) > 0 {
			poweredBy = append(poweredBy, nodeTypePowerId)
		}
	}

	return poweredBy, nil
}

// NodeCanBePoweredBy returns an array of nodes id with the type of equipment "Power",
// from which the specified node can be powered regardless of the current electricalState of the circuit breakers
func (t *TopologyGridStruct) NodeCanBePoweredBy(nodeId int) ([]int, error) {
	poweredBy := make([]int, 0)

	nodeIdx, exists := t.nodeIdxFromNodeId[nodeId]

	if !exists {
		return nil, errors.New(fmt.Sprintf("node idx was not found for node id %d", nodeId))
	}

	for _, nodeTypePowerId := range t.nodeIdArrayFromEquipmentTypeId[TypePower] {

		nodeTypePowerIdx, exists := t.nodeIdxFromNodeId[nodeTypePowerId]

		if !exists {
			return nil, errors.New(fmt.Sprintf("node idx was not found for node id %d", nodeId))
		}

		path, _ := graph.ShortestPath(t.fullGraph, nodeTypePowerIdx, nodeIdx)
		if len(path) > 0 {
			poweredBy = append(poweredBy, nodeTypePowerId)
		}
	}

	return poweredBy, nil
}

// CircuitBreakersNextToNode returns an array of circuit breakers id next to the node
func (t *TopologyGridStruct) CircuitBreakersNextToNode(nodeId int) ([]int, error) {
	var exists bool
	var nodeIdx int
	var edgeCircuitBreakerIdx int
	circuitBreakers := make([]int, 0)

	nodeIdx, exists = t.nodeIdxFromNodeId[nodeId]

	if !exists {
		return nil, errors.New(fmt.Sprintf("node idx was not found for node id %d", nodeId))
	}

	for _, edgeCircuitBreakerId := range t.edgeIdArrayFromEquipmentTypeId[TypeCircuitBreaker] {

		edgeCircuitBreakerIdx, exists = t.edgeIdxFromEdgeId[edgeCircuitBreakerId]

		if !exists {
			return nil, errors.New(fmt.Sprintf("node idx was not found for node id %d", nodeId))
		}

		circuitBreaker := t.edges[edgeCircuitBreakerIdx]

		path, pathLen := graph.ShortestPath(t.fullGraph, t.nodeIdxFromNodeId[circuitBreaker.terminal.node1Id], nodeIdx)

		if len(path) > 0 && pathLen == 0 {
			circuitBreakers = append(circuitBreakers, edgeCircuitBreakerId)
		} else {
			path, pathLen = graph.ShortestPath(t.fullGraph, t.nodeIdxFromNodeId[circuitBreaker.terminal.node2Id], nodeIdx)
			if len(path) > 0 && pathLen == 0 {
				circuitBreakers = append(circuitBreakers, edgeCircuitBreakerId)
			}
		}
	}

	return circuitBreakers, nil
}

// BfsFromNodeId traverses current graph in breadth-first order starting at nodeStart
func (t *TopologyGridStruct) BfsFromNodeId(nodeIdStart int) []TerminalStruct {

	var path []TerminalStruct

	graph.BFS(graph.Sort(t.currentGraph), t.nodeIdxFromNodeId[nodeIdStart], func(v, w int, c int64) {
		path = append(path, TerminalStruct{node1Id: t.nodes[v].id, node2Id: t.nodes[w].id, numberOfSwitches: c})
	})

	return path
}

// GetAsGraphMl returns a string with a graph represented by the graph modeling language
func (t *TopologyGridStruct) GetAsGraphMl() string {
	var graphMl string
	var graphics string

	const GraphicsPower = "\n    graphics\n    [\n      type \"star6\"\n      fill \"#FF0000\"\n    ]"
	const GraphicsConsumer = "\n    graphics\n    [\n      type \"triangle\"\n      fill \"#FFCC00\"\n    ]"
	const GraphicsJoin = "\n    graphics\n    [\n      type \"ellipse\"\n      fill \"#808080\"\n    ]"
	const GraphicsLine = "\n    graphics\n    [\n      type \"rectangle\"\n      fill \"#FF8080\"\n    ]"

	const GraphicsStateOff = "\n    graphics\n    [\n    style \"dotted\"\n      fill \"#000000\"\n    ]"
	const GraphicsCircuitBreakerOn = "\n    graphics\n    [\n    fill \"#FF0000\"\n    ]"
	const GraphicsCircuitBreakerOff = "\n    graphics\n    [\n    style \"dotted\"\n      fill \"#FF0000\"\n    ]"
	const GraphicsDisconnectSwitchOn = "\n    graphics\n    [\n    fill \"#00FF00\"\n    ]"
	const GraphicsDisconnectSwitchOff = "\n    graphics\n    [\n    style \"dotted\"\n      fill \"#00FF00\"\n    ]"

	for _, node := range t.nodes {

		if t.equipment[node.equipmentId].typeId == TypePower {
			graphics = GraphicsPower
		} else if t.equipment[node.equipmentId].typeId == TypeConsumer {
			graphics = GraphicsConsumer
		} else if t.equipment[node.equipmentId].typeId == TypeLine {
			graphics = GraphicsLine
		} else {
			graphics = GraphicsJoin
		}
		graphMl += fmt.Sprintf("  node [%s\n    id %d\n    label \"%s\"\n  ]\n",
			graphics, node.id, t.equipment[node.equipmentId].name)
	}

	for _, edge := range t.edges {
		graphics = ""

		if t.equipment[edge.equipmentId].switchState == 0 {
			graphics = GraphicsStateOff
		}

		if t.equipment[edge.equipmentId].typeId == TypeCircuitBreaker {
			if t.equipment[edge.equipmentId].switchState == 1 {
				graphics = GraphicsCircuitBreakerOn
			} else {
				graphics = GraphicsCircuitBreakerOff
			}
		} else if t.equipment[edge.equipmentId].typeId == TypeDisconnectSwitch {
			if t.equipment[edge.equipmentId].switchState == 1 {
				graphics = GraphicsDisconnectSwitchOn
			} else {
				graphics = GraphicsDisconnectSwitchOff
			}
		}

		graphMl += fmt.Sprintf("  edge [%s\n    source %d\n    target %d\n    label \"%s\"\n  ]\n",
			graphics, edge.terminal.node1Id, edge.terminal.node2Id, t.equipment[edge.equipmentId].name)
	}

	return "graph [\n" + graphMl + "]\n"
}

// SetEquipmentElectricalState for all equipment
// TODO: The electrical state of the switches (edges) in the off state must be calculated by more sophisticated algorithm, since its terminals can have different electrical states.
func (t *TopologyGridStruct) SetEquipmentElectricalState() {

	for id, equipment := range t.equipment {
		equipment.electricalState = StateIsolated
		t.equipment[id] = equipment
	}

	for idx, node := range t.nodes {
		node.electricalState = StateIsolated
		t.nodes[idx] = node
	}

	for _, nodeIdOfPowerNode := range t.nodeIdArrayFromEquipmentTypeId[TypePower] {
		cost := make(map[int]int64)

		for _, terminal := range t.BfsFromNodeId(nodeIdOfPowerNode) {
			cost[terminal.node2Id] += terminal.numberOfSwitches + cost[terminal.node1Id]

			node := t.nodes[t.nodeIdxFromNodeId[terminal.node1Id]]
			node.electricalState |= StateEnergized
			t.nodes[t.nodeIdxFromNodeId[terminal.node1Id]] = node
			if node.equipmentId != 0 {
				equipment := t.equipment[node.equipmentId]
				equipment.electricalState |= StateEnergized
				equipment.poweredBy[nodeIdOfPowerNode] = cost[terminal.node1Id]
				t.equipment[node.equipmentId] = equipment
			}

			for _, edgeId := range t.edgeIdArrayFromNodeId[node.id] {
				edge := t.edges[t.edgeIdxFromEdgeId[edgeId]]
				if edge.equipmentId != 0 {
					equipment := t.equipment[edge.equipmentId]
					equipment.electricalState |= StateEnergized
					equipment.poweredBy[nodeIdOfPowerNode] = cost[terminal.node1Id]
					t.equipment[edge.equipmentId] = equipment
				}
			}

			node = t.nodes[t.nodeIdxFromNodeId[terminal.node2Id]]
			node.electricalState |= StateEnergized
			t.nodes[t.nodeIdxFromNodeId[terminal.node2Id]] = node
			if node.equipmentId != 0 {
				equipment := t.equipment[node.equipmentId]
				equipment.electricalState |= StateEnergized
				equipment.poweredBy[nodeIdOfPowerNode] = cost[terminal.node2Id]
				t.equipment[node.equipmentId] = equipment
			}

			for _, edgeId := range t.edgeIdArrayFromNodeId[node.id] {
				edge := t.edges[t.edgeIdxFromEdgeId[edgeId]]
				if edge.equipmentId != 0 {
					equipment := t.equipment[edge.equipmentId]
					equipment.electricalState |= StateEnergized
					equipment.poweredBy[nodeIdOfPowerNode] = cost[terminal.node2Id]
					t.equipment[edge.equipmentId] = equipment
				}
			}
		}
	}
}

func (t *TopologyGridStruct) StringEquipment() {
	for _, equipment := range t.equipment {
		fmt.Printf("%4d:%30s:%2d <- %+v\n", equipment.id, equipment.name, equipment.electricalState, equipment.poweredBy)
	}
}

// GetFurthestEquipmentFromPower returns the furthest equipment from the power supply, the ID of the power supply node,
// and the number of switches between the power supply and the equipment
func (t *TopologyGridStruct) GetFurthestEquipmentFromPower(equipmentIds []int) (int, int, int64) {
	var furthestEquipmentId = 0
	var poweredByNodeId = 0

	poweredBy := make(map[int]int64)

	for _, equipmentId := range equipmentIds {
		equipment := t.equipment[equipmentId]
		if equipment.switchState == 0 {
			continue
		}
		for _poweredByNodeId, numberOfSwitches := range equipment.poweredBy {
			if poweredBy[_poweredByNodeId] < numberOfSwitches {
				poweredBy[_poweredByNodeId] = numberOfSwitches
				furthestEquipmentId = equipmentId
				poweredByNodeId = _poweredByNodeId
			}
		}
	}

	return furthestEquipmentId, poweredByNodeId, poweredBy[poweredByNodeId]
}

// GetFurthestEquipmentNodeIdFromPower returns the farthest (from two) equipment node id from the power source
func (t *TopologyGridStruct) GetFurthestEquipmentNodeIdFromPower(poweredByNodeId int, equipmentId int) int {
	var furthestNodeId = 0
	var maxNumberOfSwitches int64 = 0

	for _, nodeId := range t.nodeIdArrayFromEquipmentId[equipmentId] {
		_, numberOfSwitches := graph.ShortestPath(t.currentGraph, t.nodeIdxFromNodeId[nodeId], t.nodeIdxFromNodeId[poweredByNodeId])
		if maxNumberOfSwitches < numberOfSwitches {
			maxNumberOfSwitches = numberOfSwitches
			furthestNodeId = nodeId
		}
	}

	return furthestNodeId
}

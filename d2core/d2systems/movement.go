package d2systems

import (
	"time"

	"github.com/gravestench/akara"

	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2components"
)

// NewMovementSystem creates a movement system
func NewMovementSystem() *MovementSystem {
	cfg := akara.NewFilter().Require(d2components.Position, d2components.Velocity)

	filter := cfg.Build()

	return &MovementSystem{
		SubscriberSystem: akara.NewSubscriberSystem(filter),
	}
}

// static check that MovementSystem implements the System interface
var _ akara.System = &MovementSystem{}

// MovementSystem handles entity movement based on velocity and position components
type MovementSystem struct {
	*akara.SubscriberSystem
	positions  *d2components.PositionMap
	velocities *d2components.VelocityMap
}

// Init initializes the system with the given world
func (m *MovementSystem) Init(world *akara.World) {
	m.World = world

	if world == nil {
		m.SetActive(false)
		return
	}

	for subIdx := range m.Subscriptions {
		m.Subscriptions[subIdx] = m.AddSubscription(m.Subscriptions[subIdx].Filter)
	}

	// try to inject the components we require, then cast the returned
	// abstract ComponentMap back to the concrete implementation
	m.positions = m.InjectMap(d2components.Position).(*d2components.PositionMap)
	m.velocities = m.InjectMap(d2components.Velocity).(*d2components.VelocityMap)
}

// Process processes all of the Entities
func (m *MovementSystem) Process() {
	for subIdx := range m.Subscriptions {
		entities := m.Subscriptions[subIdx].GetEntities()
		for entIdx := range entities {
			m.ProcessEntity(entities[entIdx])
		}
	}
}

// ProcessEntity updates an individual entity in the movement system
func (m *MovementSystem) ProcessEntity(id akara.EID) {
	position, found := m.positions.GetPosition(id)
	if !found {
		return
	}

	velocity, found := m.velocities.GetVelocity(id)
	if !found {
		return
	}

	s := float64(m.World.TimeDelta) / float64(time.Second)
	position.Vector = *position.Vector.Add(velocity.Vector.Clone().Scale(s))
}
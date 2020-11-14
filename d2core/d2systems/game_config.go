package d2systems

import (
	"encoding/json"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"

	"github.com/gravestench/akara"

	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2components"
)

// static check that the game config system implements the system interface
var _ akara.System = &GameConfigSystem{}

func NewGameConfigSystem() *GameConfigSystem {
	// we are going to check entities that dont yet have loaded asset types
	thingsToCheck := akara.NewFilter().
		Require(d2components.FilePath).
		Require(d2components.FileType).
		Require(d2components.FileHandle).
		Forbid(d2components.GameConfig).
		Forbid(d2components.StringTable).
		Forbid(d2components.DataDictionary).
		Forbid(d2components.Palette).
		Forbid(d2components.PaletteTransform).
		Forbid(d2components.Cof).
		Forbid(d2components.Dc6).
		Forbid(d2components.Dcc).
		Forbid(d2components.Ds1).
		Forbid(d2components.Dt1).
		Forbid(d2components.Wav).
		Forbid(d2components.AnimData).
		Build()

	// we are interested in actual game config instances, too
	gameConfigs := akara.NewFilter().Require(d2components.GameConfig).Build()

	gcs := &GameConfigSystem{
		SubscriberSystem: akara.NewSubscriberSystem(thingsToCheck, gameConfigs),
		maps: struct {
			gameConfigs *d2components.GameConfigMap
			filePaths   *d2components.FilePathMap
			fileTypes   *d2components.FileTypeMap
			fileHandles *d2components.FileHandleMap
			fileSources *d2components.FileSourceMap
			dirty       *d2components.DirtyMap
		}{},
	}

	return gcs
}

// GameConfigSystem is responsible for game config bootstrap procedure, as well as
// clearing the `Dirty` component of game configs. In the `bootstrap` method of this system
// you can see that this system will add entities for the directories it expects config files
// to be found in, and it also adds an entity for the initial config file to be loaded.
//
// This system is dependant on the FileTypeResolver, FileSourceResolver, and
// FileHandleResolver systems because this system subscribes to entities
// with components created by these other systems. Nothing will  break if these
// other systems are not present in the world, but no config files will be loaded by
// this system either...
type GameConfigSystem struct {
	*akara.SubscriberSystem
	filesToCheck *akara.Subscription
	gameConfigs  *akara.Subscription
	maps         struct {
		gameConfigs *d2components.GameConfigMap
		filePaths   *d2components.FilePathMap
		fileTypes   *d2components.FileTypeMap
		fileHandles *d2components.FileHandleMap
		fileSources *d2components.FileSourceMap
		dirty       *d2components.DirtyMap
	}
}

func (m *GameConfigSystem) Init(world *akara.World) {
	m.World = world

	if world == nil {
		m.SetActive(false)
		return
	}

	for subIdx := range m.Subscriptions {
		m.Subscriptions[subIdx] = m.AddSubscription(m.Subscriptions[subIdx].Filter)
	}

	m.filesToCheck = m.Subscriptions[0]
	m.gameConfigs = m.Subscriptions[1]

	// try to inject the components we require, then cast the returned
	// abstract ComponentMap back to the concrete implementation
	m.maps.filePaths = world.InjectMap(d2components.FilePath).(*d2components.FilePathMap)
	m.maps.fileTypes = world.InjectMap(d2components.FileType).(*d2components.FileTypeMap)
	m.maps.fileHandles = world.InjectMap(d2components.FileHandle).(*d2components.FileHandleMap)
	m.maps.fileSources = world.InjectMap(d2components.FileSource).(*d2components.FileSourceMap)
	m.maps.gameConfigs = world.InjectMap(d2components.GameConfig).(*d2components.GameConfigMap)
	m.maps.dirty = world.InjectMap(d2components.Dirty).(*d2components.DirtyMap)
}

func (m *GameConfigSystem) Process() {
	m.clearDirty(m.gameConfigs.GetEntities())
	m.checkForNewConfig(m.filesToCheck.GetEntities())
}

func (m *GameConfigSystem) clearDirty(entities []akara.EID) {
	for _, eid := range entities {
		dc, found := m.maps.dirty.GetDirty(eid)
		if !found {
			m.maps.dirty.AddDirty(eid) // adds it, but it's false
			continue
		}

		dc.IsDirty = false
	}
}

func (m *GameConfigSystem) checkForNewConfig(entities []akara.EID) {
	for _, eid := range entities {
		fp, found := m.maps.filePaths.GetFilePath(eid)
		if !found {
			continue
		}

		ft, found := m.maps.fileTypes.GetFileType(eid)
		if !found {
			continue
		}

		if fp.Path != configFileName || ft.Type != d2enum.FileTypeJSON {
			continue
		}

		m.loadConfig(eid)
	}
}

func (m *GameConfigSystem) loadConfig(eid akara.EID) {
	fh, found := m.maps.fileHandles.GetFileHandle(eid)
	if !found {
		return
	}

	gameConfig := m.maps.gameConfigs.AddGameConfig(eid)

	if err := json.NewDecoder(fh.Data).Decode(gameConfig); err != nil {
		m.maps.gameConfigs.Remove(eid)
		return
	}
}
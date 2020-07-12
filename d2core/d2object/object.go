// Package d2object implements objects placed on the map and their functionality
package d2object

import (
	"math"
	"math/rand"

	"github.com/OpenDiablo2/OpenDiablo2/d2common"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2data/d2datadict"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2interface"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2resource"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2asset"
)

// Object represents a composite of animations that can be projected onto the map.
type Object struct {
	composite          d2interface.CompositeAnimation
	highlight          bool
	LocationX          float64
	LocationY          float64
	TileX, TileY       int     // Coordinates of the tile the unit is within
	subcellX, subcellY float64 // Subcell coordinates within the current tile
	// nameLabel    d2ui.Label
	objectRecord *d2datadict.ObjectRecord
	drawLayer    int
	name         string
}

// CreateObject creates an instance of AnimatedComposite
func CreateObject(x, y int, objectRec *d2datadict.ObjectRecord, palettePath string) (*Object, error) {
	locX, locY := float64(x), float64(y)
	entity := &Object{
		objectRecord: objectRec,
		LocationX:    locX,
		LocationY:    locY,
		subcellX:     1 + math.Mod(locX, 5),
		subcellY:     1 + math.Mod(locY, 5),
		TileX:        x / 5,
		TileY:        y / 5,
		name:         d2common.TranslateString(objectRec.Name),
	}
	objectType := &d2datadict.ObjectTypes[objectRec.Index]

	// TODO get ri of the singleton asset manager
	composite, err := d2asset.LoadComposite(d2enum.ObjectTypeItem, objectType.Token,
		d2resource.PaletteUnits)
	if err != nil {
		return nil, err
	}

	entity.composite = composite

	entity.setMode(d2enum.ObjectAnimationModeNeutral, 0, false)

	initObject(entity)

	return entity, nil
}

// setMode changes the graphical mode of this animated entity
func (ob *Object) setMode(animationMode d2enum.ObjectAnimationMode, direction int, randomFrame bool) error {
	err := ob.composite.SetMode(animationMode, "HTH")
	if err != nil {
		return err
	}

	ob.composite.SetDirection(direction)

	// mode := d2enum.ObjectAnimationModeFromString(animationMode)
	ob.drawLayer = ob.objectRecord.OrderFlag[d2enum.ObjectAnimationModeNeutral]

	// For objects their txt record entry overrides animationdata
	speed := ob.objectRecord.FrameDelta[animationMode]
	if speed != 0 {
		ob.composite.SetAnimSpeed(speed)
	}

	frameCount := ob.objectRecord.FrameCount[animationMode]

	if frameCount != 0 {
		ob.composite.SetSubLoop(0, frameCount)
	}

	ob.composite.SetPlayLoop(ob.objectRecord.CycleAnimation[animationMode])
	ob.composite.SetCurrentFrame(ob.objectRecord.StartFrame[animationMode])

	if randomFrame {
		n := rand.Intn(frameCount)
		ob.composite.SetCurrentFrame(n)
	}

	return err
}

// Highlight sets the entity highlighted flag to true.
func (ob *Object) Highlight() {
	ob.highlight = true
}

func (ob *Object) Selectable() bool {
	modeString := ob.composite.GetAnimationMode().String()
	mode := d2enum.ObjectAnimationModeFromString(modeString)
	return ob.objectRecord.Selectable[mode]
}

// Render draws this animated entity onto the target
func (ob *Object) Render(target d2interface.Surface) {
	target.PushTranslation(
		int((ob.subcellX-ob.subcellY)*16),
		int(((ob.subcellX + ob.subcellY) * 8)),
	)

	if ob.highlight {
		target.PushBrightness(2)
		defer target.Pop()
	}

	defer target.Pop()
	ob.composite.Render(target)
	ob.highlight = false
}

// Advance updates the animation
func (ob *Object) Advance(elapsed float64) {
	ob.composite.Advance(elapsed)
}

// GetLayer returns which layer of the map the object is drawn
func (ob *Object) GetLayer() int {
	return ob.drawLayer
}

// GetPosition of the object
func (ob *Object) GetPosition() (x, y float64) {
	return float64(ob.TileX), float64(ob.TileY)
}

// GetPositionF of the object but differently
func (ob *Object) GetPositionF() (x, y float64) {
	return float64(ob.TileX) + (ob.subcellX / 5.0), float64(ob.TileY) + ob.subcellY/5.0
}

// Name gets the name of the object
func (ob *Object) Name() string {
	return ob.name
}

package mir

import (
	"fmt"
	"sync"
	"time"

	"github.com/yenkeia/mirgo/common"
	"github.com/yenkeia/mirgo/proto/server"
)

type IBehavior interface {
	Process()
}

type BehaviroFactory func(id int, mon *Monster) IBehavior

var behaviorFactory BehaviroFactory

func SetMonsterBehaviorFactory(fac BehaviroFactory) {
	behaviorFactory = fac
}

// Monster ...
type Monster struct {
	MapObject
	Image       common.Monster
	AI          int
	Behavior    IBehavior
	Effect      int
	Poison      common.PoisonType
	Light       uint8
	Target      IMapObject
	Level       uint16
	PetLevel    uint16
	Experience  uint16
	HP          uint32
	MaxHP       uint32
	MinAC       uint16
	MaxAC       uint16
	MinMAC      uint16
	MaxMAC      uint16
	MinDC       uint16
	MaxDC       uint16
	MinMC       uint16
	MaxMC       uint16
	MinSC       uint16
	MaxSC       uint16
	Accuracy    uint8
	Agility     uint8
	MoveSpeed   uint16
	AttackSpeed int32
	ArmourRate  float32
	DamageRate  float32
	ViewRange   int
	Master      *Player
	EXPOwner    *Player
	ActionList  *sync.Map // map[uint32]DelayedAction
	ActionTime  time.Time
	AttackTime  time.Time
	DeadTime    time.Time
	MoveTime    time.Time
}

func (m *Monster) String() string {
	return fmt.Sprintf("Monster: %s, (%v), ID: %d, ptr: %p\n", m.Name, m.CurrentLocation, m.ID, m)
}

// NewMonster ...
func NewMonster(mp *Map, p common.Point, mi *common.MonsterInfo) (m *Monster) {
	m = new(Monster)
	m.ID = mp.Env.NewObjectID()
	m.Map = mp
	m.Name = mi.Name
	m.NameColor = common.Color{R: 255, G: 255, B: 255}
	m.Image = common.Monster(mi.Image)
	m.AI = mi.AI
	m.Effect = mi.Effect
	m.Light = uint8(mi.Light)
	m.Target = nil
	m.Poison = common.PoisonTypeNone
	m.CurrentLocation = p
	m.CurrentDirection = RandomDirection()
	m.Dead = false
	m.Level = uint16(mi.Level)
	m.PetLevel = 0
	m.Experience = uint16(mi.Experience)
	m.HP = uint32(mi.HP)
	m.MaxHP = uint32(mi.HP)
	m.MinAC = uint16(mi.MinAC)
	m.MaxAC = uint16(mi.MaxAC)
	m.MinMAC = uint16(mi.MinMAC)
	m.MaxMAC = uint16(mi.MaxMAC)
	m.MinDC = uint16(mi.MinDC)
	m.MaxDC = uint16(mi.MaxDC)
	m.MinMC = uint16(mi.MinMC)
	m.MaxMC = uint16(mi.MaxMC)
	m.MinSC = uint16(mi.MinSC)
	m.MaxSC = uint16(mi.MaxSC)
	m.Accuracy = uint8(mi.Accuracy)
	m.Agility = uint8(mi.Agility)
	m.MoveSpeed = uint16(mi.MoveSpeed)
	m.AttackSpeed = int32(mi.AttackSpeed)
	m.ArmourRate = 1.0
	m.DamageRate = 1.0
	m.ActionList = new(sync.Map)
	now := time.Now()
	m.ActionTime = now
	m.MoveTime = now
	m.ViewRange = mi.ViewRange
	m.Behavior = behaviorFactory(m.AI, m)
	return m
}

func (m *Monster) GetID() uint32 {
	return m.ID
}

func (m *Monster) GetName() string {
	return m.Name
}

func (m *Monster) GetRace() common.ObjectType {
	return common.ObjectTypeMonster
}

func (m *Monster) GetPoint() common.Point {
	return m.CurrentLocation
}

func (m *Monster) GetCell() *Cell {
	return m.Map.GetCell(m.CurrentLocation)
}

func (m *Monster) GetDirection() common.MirDirection {
	return m.CurrentDirection
}

func (m *Monster) GetInfo() interface{} {
	res := &server.ObjectMonster{
		ObjectID:          m.ID,
		Name:              m.Name,
		NameColor:         m.NameColor.ToInt32(),
		Location:          m.GetPoint(),
		Image:             m.Image,
		Direction:         m.GetDirection(),
		Effect:            uint8(m.Effect),
		AI:                uint8(m.AI),
		Light:             m.Light,
		Dead:              m.IsDead(),
		Skeleton:          m.IsSkeleton(),
		Poison:            m.Poison,
		Hidden:            m.IsHidden(),
		ShockTime:         0,     // TODO
		BindingShotCenter: false, // TODO
		Extra:             false, // TODO
		ExtraByte:         0,     // TODO
	}
	return res
}

func (m *Monster) GetBaseStats() BaseStats {
	return BaseStats{
		MinAC:    m.MinAC,
		MaxAC:    m.MaxAC,
		MinMAC:   m.MinMAC,
		MaxMAC:   m.MaxMAC,
		MinDC:    m.MinDC,
		MaxDC:    m.MaxDC,
		MinMC:    m.MinMC,
		MaxMC:    m.MaxMC,
		MinSC:    m.MinSC,
		MaxSC:    m.MaxSC,
		Accuracy: m.Accuracy,
		Agility:  m.Agility,
	}
}

func (m *Monster) AddBuff(buff *Buff) {}

func (m *Monster) ApplyPoison(poison *Poison, caster IMapObject) {}

func (m *Monster) Broadcast(msg interface{}) {
	m.Map.BroadcastP(m.CurrentLocation, msg, nil)
}

// Spawn 怪物生成
func (m *Monster) Spawn(mp *Map, p common.Point) {
	m.Map = mp
	m.CurrentLocation = p
	mp.AddObject(m)
}

func (m *Monster) BroadcastDamageIndicator(typ common.DamageType, dmg int) {
	m.Broadcast(ServerMessage{}.DamageIndicator(int32(dmg), typ, m.GetID()))
}

func (m *Monster) IsDead() bool {
	return m.Dead
}

func (m *Monster) IsUndead() bool {
	return false
}

func (m *Monster) IsBlocking() bool {
	return !m.IsDead()
}

func (m *Monster) IsSkeleton() bool {
	return false
}

func (m *Monster) IsHidden() bool {
	return false
}

func (m *Monster) IsAttackTargetMonster(attacker *Monster) bool {
	if attacker == m {
		return false
	}

	if m.AI == 6 || m.AI == 58 {
		return false
	}

	if attacker.AI == 6 {
		if m.AI != 1 && m.AI != 2 && m.AI != 3 { //Not Dear/Hen/Tree/Pets or Red Master
			return true
		}
	} else if attacker.AI == 58 {
		if m.AI != 1 && m.AI != 2 && m.AI != 3 {
			return true
		}
	}
	return false
}

func (m *Monster) IsAttackTarget(attacker IMapObject) bool {

	switch attacker.(type) {
	case *Monster:
		return m.IsAttackTargetMonster(attacker.(*Monster))
	case *Player:
	}
	return true
}

func (m *Monster) IsFriendlyTarget(attacker IMapObject) bool {
	return false
}

func (m *Monster) AttackMode() common.AttackMode {
	return common.AttackModeAll
}

func (m *Monster) CanMove() bool {
	return time.Now().After(m.MoveTime)
}

func (m *Monster) CanAttack() bool {
	now := time.Now()
	if m.IsDead() {
		return false
	}
	return now.After(m.AttackTime)
}

// InAttackRange 是否在怪物攻击范围内
func (m *Monster) InAttackRange() bool {
	// if (Target.CurrentMap != CurrentMap) return false;
	return !m.Target.GetPoint().Equal(m.CurrentLocation) && InRange(m.CurrentLocation, m.Target.GetPoint(), 1)
}

// Process 怪物定时轮询
func (m *Monster) Process() {
	if m.Target != nil &&
		//m.Target.GetMap() != m.Map ||
		(!m.Target.IsAttackTarget(m) || !InRange(m.CurrentLocation, m.Target.GetPoint(), DataRange)) {
		m.Target = nil
	}

	now := time.Now()

	if m.IsDead() && m.DeadTime.Before(now) {
		m.Map.DeleteObject(m)
		m.Broadcast(&server.ObjectRemove{ObjectID: m.GetID()})
		return
	}

	m.Behavior.Process()

	m.ProcessBuffs()
	m.ProcessRegan()
	m.ProcessPoison()

	finishID := make([]uint32, 0)
	m.ActionList.Range(func(k, v interface{}) bool {
		action := v.(*DelayedAction)
		if action.Finish || now.Before(action.ActionTime) {
			return true
		}
		action.Task.Execute()
		action.Finish = true
		if action.Finish {
			finishID = append(finishID, action.ID)
		}
		return true
	})
	for i := range finishID {
		m.ActionList.Delete(finishID[i])
	}
}

// ProcessBuffs 处理怪物增益效果
func (m *Monster) ProcessBuffs() {

}

// ProcessRegan 怪物自身回血
func (m *Monster) ProcessRegan() {

}

// ProcessPoison 处理怪物中毒效果
func (m *Monster) ProcessPoison() {

}

// GetDefencePower 获取防御值
func (m *Monster) GetDefencePower(min, max int) int {
	if min < 0 {
		min = 0
	}
	if min > max {
		max = min
	}
	return RandomInt(min, max)
}

// GetAttackPower 获取攻击值
func (m *Monster) GetAttackPower(min, max int) int {
	if min < 0 {
		min = 0
	}
	if min > max {
		max = min
	}
	// TODO luck
	return RandomInt(min, max+1)
}

// Die ...
func (m *Monster) Die() {
	if m.IsDead() {
		return
	}

	m.HP = 0
	m.Dead = true
	m.DeadTime = time.Now().Add(5 * time.Second)

	m.Broadcast(ServerMessage{}.ObjectDied(m.GetID(), m.GetDirection(), m.GetPoint()))
	// EXPOwner.WinExp(Experience, Level);

	if m.EXPOwner != nil && m.Master == nil && m.EXPOwner.GetRace() == common.ObjectTypePlayer {
		m.EXPOwner.WinExp(int(m.Experience), int(m.Level))
		// PlayerObject playerObj = (PlayerObject)EXPOwner;
		// playerObj.CheckGroupQuestKill(Info);
		// m.EXPOwner.CheckGroupQuestKill(Info)
	}

	m.Drop()
}

// ChangeHP 怪物改变血量 amount 可以是负数(扣血)
func (m *Monster) ChangeHP(amount int) {
	if m.IsDead() {
		return
	}
	value := int(m.HP) + amount
	if value == int(m.HP) {
		return
	}
	if value <= 0 {
		m.Die()
		m.HP = 0
	} else {
		m.HP = uint32(value)
	}
	percent := uint8(float32(m.HP) / float32(m.MaxHP) * 100)
	log.Debugf("monster HP: %d, MaxHP: %d, percent: %d\n", m.HP, m.MaxHP, percent)
	m.Broadcast(ServerMessage{}.ObjectHealth(m.GetID(), percent, 5))
}

// Attacked 被攻击
func (m *Monster) Attacked(attacker IMapObject, damage int, defenceType common.DefenceType, damageWeapon bool) {
	if m.Target == nil && attacker.IsAttackTarget(m) {
		m.Target = attacker
	}
	armor := 0
	switch defenceType {
	case common.DefenceTypeACAgility:
		if RandomInt(0, int(m.Agility)) > int(attacker.GetBaseStats().Accuracy) {
			m.BroadcastDamageIndicator(common.DamageTypeMiss, 0)
			return
		}
		armor = m.GetDefencePower(int(m.MinAC), int(m.MaxAC))
	case common.DefenceTypeAC:
		armor = m.GetDefencePower(int(m.MinAC), int(m.MaxAC))
	case common.DefenceTypeMACAgility:
		if RandomInt(0, int(m.Agility)) > int(attacker.GetBaseStats().Accuracy) {
			m.BroadcastDamageIndicator(common.DamageTypeMiss, 0)
			return
		}
		armor = m.GetDefencePower(int(m.MinMAC), int(m.MaxMAC))
	case common.DefenceTypeMAC:
		armor = m.GetDefencePower(int(m.MinMAC), int(m.MaxMAC))
	case common.DefenceTypeAgility:
		if RandomInt(0, int(m.Agility)) > int(attacker.GetBaseStats().Accuracy) {
			m.BroadcastDamageIndicator(common.DamageTypeMiss, 0)
			return
		}
	}
	armor = int(float32(armor) * m.ArmourRate)
	damage = int(float32(damage) * m.DamageRate)
	value := damage - armor
	log.Debugf("attacker damage: %d, monster armor: %d\n", damage, armor)
	if value <= 0 {
		m.BroadcastDamageIndicator(common.DamageTypeMiss, 0)
		return
	}
	// TODO 还有很多没做
	m.Broadcast(ServerMessage{}.ObjectStruck(m, attacker.GetID()))
	m.BroadcastDamageIndicator(common.DamageTypeHit, value)
	m.ChangeHP(-value)
	log.Debugf("!!!attacker damage: %d, monster armor: %d\n", damage, armor)
}

// Drop 怪物掉落物品
func (m *Monster) Drop() {
	value, ok := m.Map.Env.GameDB.DropInfoMap.Load(m.Name)
	if !ok {
		return
	}
	dropInfos := value.([]common.DropInfo)
	mapItems := make([]Item, 0)
	for i := range dropInfos {
		drop := dropInfos[i]
		if RandomInt(1, drop.Chance) != 1 {
			continue
		}
		if drop.Gold > 0 {
			mapItems = append(mapItems, Item{
				MapObject: MapObject{
					ID:  m.Map.Env.NewObjectID(),
					Map: m.Map,
				},
				Gold:     uint64(drop.Gold),
				UserItem: nil,
			})
			continue
		}
		info := m.Map.Env.GameDB.GetItemInfoByName(drop.ItemName)
		if info == nil {
			continue
		}
		mapItems = append(mapItems, Item{
			MapObject: MapObject{
				ID:  m.Map.Env.NewObjectID(),
				Map: m.Map,
			},
			Gold:     0,
			UserItem: m.Map.Env.NewUserItem(info),
		})
	}
	for i := range mapItems {
		if msg, ok := mapItems[i].Drop(m.GetPoint(), 3); !ok {
			log.Warnln(msg)
		}
	}
}

// Walk 移动，成功返回 true
func (m *Monster) Walk(dir common.MirDirection) bool {
	if !m.CanMove() {
		return false
	}

	dest := m.CurrentLocation.NextPoint(dir, 1)
	destcell := m.Map.GetCell(dest)

	if destcell != nil && destcell.Objects != nil {
		blocking := false
		destcell.Objects.Range(func(_, v interface{}) bool {
			o := v.(IMapObject)
			if o.IsBlocking() || m.GetRace() == common.ObjectTypeCreature {
				blocking = true
				return false
			}
			return true
		})
		if blocking {
			return false
		}
	} else {
		return false
	}

	m.Map.GetCell(m.CurrentLocation).DeleteObject(m)
	destcell.AddObject(m)

	oldpos := m.CurrentLocation

	m.CurrentDirection = dir
	m.CurrentLocation = dest

	m.WalkNotify(oldpos, destcell.Point)

	m.MoveTime = m.MoveTime.Add(time.Duration(int64(m.MoveSpeed)) * time.Millisecond)

	m.Broadcast(&server.ObjectWalk{
		ObjectID:  m.GetID(),
		Direction: dir,
		Location:  dest,
	})

	return true
}

func (m *Monster) WalkNotify(from, to common.Point) {
	cells := m.Map.CalcDiff(from, to, DataRange)
	for c, isadd := range cells.M {
		if isadd {
			c.Objects.Range(func(k, v interface{}) bool {
				switch v.(type) {
				case *Player:
					v.(*Player).Enqueue(ServerMessage{}.Object(m))
				}

				return true
			})
		} else {
			c.Objects.Range(func(k, v interface{}) bool {
				switch v.(type) {
				case *Player:
					v.(*Player).Enqueue(ServerMessage{}.ObjectRemove(m))
				}
				return true
			})
		}

	}
}

func (m *Monster) Turn(dir common.MirDirection) {
	if !m.CanMove() {
		return
	}
	m.CurrentDirection = dir

	m.Broadcast(&server.ObjectTurn{
		ObjectID:  m.GetID(),
		Direction: dir,
		Location:  m.CurrentLocation,
	})

	// TODO:
	// InSafeZone = CurrentMap.GetSafeZone(CurrentLocation) != null

	// Cell cell = CurrentMap.GetCell(CurrentLocation);
	// for (int i = 0; i < cell.Objects.Count; i++)
	// {
	//     if (cell.Objects[i].Race != ObjectType.Spell) continue;
	//     SpellObject ob = (SpellObject)cell.Objects[i];

	//     ob.ProcessSpell(this);
	//     //break;
	// }

}

func ObjectBack(m IMapObject) common.Point {
	return m.GetPoint().NextPoint(m.GetDirection(), 1)
}

// 专用于大刀卫士攻击
func (m *Monster) GuardAttack() {
	if !m.Target.IsAttackTarget(m) {
		return
	}

	target := ObjectBack(m.Target)

	m.CurrentDirection = DirectionFromPoint(target, m.Target.GetPoint())

	m.Broadcast(&server.ObjectAttack{
		ObjectID:  m.GetID(),
		LocationX: int32(target.X),
		LocationY: int32(target.Y),
		Direction: m.CurrentDirection,
		Spell:     common.SpellNone,
		Level:     uint8(0),
		Type:      uint8(0),
	})
	m.Broadcast(&server.ObjectTurn{
		ObjectID:  m.GetID(),
		Direction: m.CurrentDirection,
		Location:  m.CurrentLocation,
	})

	now := time.Now()
	// ActionTime = Envir.Time + 300;
	m.AttackTime = now.Add(time.Duration(m.AttackSpeed) * time.Millisecond)

	damage := m.GetAttackPower(int(m.MinDC), int(m.MaxDC))

	if m.Target.GetRace() == common.ObjectTypePlayer {
		damage = int(^uint(0) >> 1) // INTMAX
	}

	if damage <= 0 {
		return
	}

	switch m.Target.GetRace() {
	case common.ObjectTypePlayer:
		m.Target.(*Player).Attacked(m, damage, common.DefenceTypeAgility, false)
	case common.ObjectTypeMonster:
		m.Target.(*Monster).Attacked(m, damage, common.DefenceTypeAgility, false)
	}
}

func (m *Monster) Attack() {
	if !m.Target.IsAttackTarget(m) {
		m.Target = nil
		return
	}
	m.CurrentDirection = DirectionFromPoint(m.CurrentLocation, m.Target.GetPoint())
	m.Broadcast(ServerMessage{}.ObjectAttack(m, common.SpellNone, 0, 0))
	now := time.Now()
	// ActionTime = Envir.Time + 300;
	m.AttackTime = now.Add(time.Duration(m.AttackSpeed) * time.Millisecond)
	damage := m.GetAttackPower(int(m.MinDC), int(m.MaxDC))
	if damage <= 0 {
		return
	}
	switch m.Target.GetRace() {
	case common.ObjectTypePlayer:
		m.Target.(*Player).Attacked(m, damage, common.DefenceTypeAgility, false)
	case common.ObjectTypeMonster:
		m.Target.(*Monster).Attacked(m, damage, common.DefenceTypeAgility, false)
	}
}

func (m *Monster) MoveTo(location common.Point) {
	if m.CurrentLocation.Equal(location) {
		return
	}
	inRange := InRange(location, m.CurrentLocation, 1)
	if inRange {
		cell := m.Map.GetCell(location)
		if cell == nil || !cell.IsValid() {
			return
		}
		ret := false
		cell.Objects.Range(func(f, v interface{}) bool {
			o := v.(IMapObject)
			if !o.IsBlocking() {
				return true
			}
			ret = true
			return false
		})
		if ret {
			return
		}
	}
	dir := DirectionFromPoint(m.CurrentLocation, location)
	if m.Walk(dir) {
		return
	}
	switch RandomNext(2) { //No favour
	case 0:
		for i := 0; i < 7; i++ {
			dir = NextDirection(dir)
			if m.Walk(dir) {
				return
			}
		}
	default:
		for i := 0; i < 7; i++ {
			dir = PreviousDirection(dir)
			if m.Walk(dir) {
				return
			}
		}
	}
}

// FindTarget 怪物寻找攻击目标
func (m *Monster) FindTarget() {
	m.Map.RangeObject(m.CurrentLocation, m.ViewRange, func(o IMapObject) bool {

		if o == m {
			return true
		}

		switch o.GetRace() {
		case common.ObjectTypeMonster:
			if !o.IsAttackTarget(m) {
				return true
			}
			// if (ob.Hidden && (!CoolEye || Level < ob.Level)) continue;
			m.Target = o

		case common.ObjectTypePlayer:

			if !o.IsAttackTarget(m) { // continue
				return true
			}

			// TODO:
			// if (playerob.GMGameMaster || ob.Hidden && (!CoolEye || Level < ob.Level) || Envir.Time < HallucinationTime) continue;

			m.Target = o

			return false
		}

		return true
	})
}

func (m *Monster) CheckStacked() bool {
	cell := m.Map.GetCell(m.CurrentLocation)
	if cell != nil && cell.Objects != nil {
		ret := false
		cell.Objects.Range(func(k, v interface{}) bool {
			ob := v.(IMapObject)
			if ob == m || ob.IsBlocking() {
				ret = true
			}
			return ret
		})
	}

	return false
}

// PetRecall 宠物传送回玩家身边
func (m *Monster) PetRecall(...interface{}) {
	log.Debugln("PetRecall", m.GetID())
}

package prattle

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrattleWithMoreThanOneNode(t *testing.T) {
	prattleOne, errOne := NewPrattle("0.0.0.0:9000, 0.0.0.0:9001", 9000)
	prattleTwo, errTwo := NewPrattle("0.0.0.0:9000, 0.0.0.0:9001", 9001)
	assert.Nil(t, errOne)
	assert.Nil(t, errTwo)
	assert.Equal(t, prattleOne.Members(), prattleTwo.Members())
	assert.Equal(t, prattleOne.members.LocalNode().Port, 9000)
	assert.Equal(t, prattleTwo.members.LocalNode().Port, 9001)
	assert.Equal(t, 2, prattleOne.members.NumMembers())
	assert.Equal(t, 2, prattleOne.broadcasts.NumNodes())
	assert.Equal(t, 3, prattleOne.broadcasts.RetransmitMult)
	defer prattleOne.Shutdown()
	defer prattleTwo.Shutdown()
}

func TestNewPrattleWhenMemberAddressIsNotInUse(t *testing.T) {
	prattle, err := NewPrattle("", 9000)
	assert.Nil(t, err)
	assert.Equal(t, 1, prattle.members.NumMembers())
	assert.Equal(t, 1, prattle.broadcasts.NumNodes())
	assert.NotNil(t, prattle.database.connection)
	defer prattle.Shutdown()
}

func TestPrattleWhenMemberAddressIsAlreadyInUse(t *testing.T) {
	prattle, errOne := NewPrattle("", 9000)
	_, errTwo := NewPrattle("", 9000)
	assert.Nil(t, errOne)
	assert.NotNil(t, errTwo)
	defer prattle.Shutdown()
}

func TestGetWhenKeyIsNotFound(t *testing.T) {
	prattle, _ := NewPrattle("", 9000)
	value, found := prattle.Get("ping")
	assert.False(t, found)
	assert.Equal(t, value, nil)
	defer prattle.Shutdown()
}

func TestGetWhenKeyIsFound(t *testing.T) {
	prattle, _ := NewPrattle("", 9000)
	prattle.Set("ping", "pong")
	value, found := prattle.Get("ping")
	assert.True(t, found)
	assert.Equal(t, "pong", value)
	defer prattle.Shutdown()
}

func TestSetWhenKeyAlreadyExist(t *testing.T) {
	prattle, _ := NewPrattle("", 9000)
	prattle.Set("ping", "pong")
	value, _ := prattle.Get("ping")
	assert.Equal(t, "pong", value)
	prattle.Set("ping", "pong2")
	newValue, _ := prattle.Get("ping")
	assert.Equal(t, "pong2", newValue)
	defer prattle.Shutdown()
}

func TestMembers(t *testing.T) {
	prattleOne, _ := NewPrattle("0.0.0.0:9000,0.0.0.0:9001", 9000)
	prattleTwo, _ := NewPrattle("0.0.0.0:9000,0.0.0.0:9001", 9001)
	assert.Equal(t, 2, len(prattleOne.Members()))
	prattleTwo.Shutdown()
	prattleOne.Shutdown()
}

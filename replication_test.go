package prattle

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemberListWhenThereIsOnlyOneMember(t *testing.T) {
	memberList, err := newMemberlist(9000, "", nil)
	defer memberList.Shutdown()
	node := memberList.LocalNode()
	require.NoError(t, err)
	assert.Equal(t, 1, memberList.NumMembers())
	assert.Equal(t, uint16(9000), node.Port)
}

func TestNewMemberListWhenThereAreMoreThanOneMembers(t *testing.T) {
	memberListOne, err := newMemberlist(9000, "", nil)
	memberListTwo, err := newMemberlist(9001, "0.0.0.0:9000,0.0.0.0:9001", nil)
	defer memberListOne.Shutdown()
	defer memberListTwo.Shutdown()
	node := memberListTwo.LocalNode()
	require.NoError(t, err)
	assert.Equal(t, 2, memberListTwo.NumMembers())
	assert.Equal(t, uint16(9001), node.Port)
}

func TestNewMemberListWhenAddressIsAlreadyInUse(t *testing.T) {
	memberListOne, errOne := newMemberlist(9001, "", nil)
	_, errTwo := newMemberlist(9001, "", nil)
	defer memberListOne.Shutdown()
	require.NoError(t, errOne)
	require.Error(t, errTwo)
}

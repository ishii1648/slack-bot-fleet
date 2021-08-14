package route

import (
	"reflect"
	"testing"
)

func TestParseReactionAddedItem(t *testing.T) {
	items, err := ParseReactionAddedItem("../../tests/routing.yml")
	if err != nil {
		t.Fatal(err)
	}

	want := []ReactionAddedItem{
		{ServiceName: "test", Type: "reaction_added", Users: []string{"s_ishii"}, Reactions: []string{"ok_hand"}, ItemChannels: []string{"development"}},
		{ServiceName: "test02", Type: "reaction_added", Users: []string{"m_taro"}, Reactions: []string{"ok_hand"}, ItemChannels: []string{"sample"}},
	}
	if want, got := want, items; !reflect.DeepEqual(want, got) {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestReactionAddedItemMatch(t *testing.T) {
	item := &ReactionAddedItem{
		ServiceName:  "test",
		Type:         "reaction_added",
		Users:        []string{"s_ishii"},
		Reactions:    []string{"ok_hand"},
		ItemChannels: []string{"development"},
	}

	if want, got := true, item.Match("s_ishii", "ok_hand", "development"); want != got {
		t.Errorf("want %v, got %v", want, got)
	}
}

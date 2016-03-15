package telemetry

import "testing"

func TestSubcontext(t *testing.T) {
	parentTags := make([]string, 1, 3)
	parentTags[0] = "a"

	parent := Context{tags: parentTags}
	child1 := parent.SubContext("b")
	child2 := parent.SubContext("c")

	if child1.Tags()[0] != "a" {
		t.Error("wrong tag on child1.tags[0]", child1.Tags()[0])
	}

	if child1.Tags()[1] != "b" {
		t.Error("wrong tag on child1.tags[1]", child1.Tags()[1])
	}

	if child2.Tags()[0] != "a" {
		t.Error("wrong tag on child2.tags[0]", child2.Tags()[0])
	}

	if child2.Tags()[1] != "c" {
		t.Error("wrong tag on child2.tags[1]", child2.Tags()[1])
	}
}

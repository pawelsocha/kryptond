package mikrotik

import (
	"testing"
)

func TestDevicePrint(t *testing.T) {
	//given
	client := &ClientMock{}
	entity := Queue{
		Name: "test",
	}

	sentences := []string{
		"/queue/simple/print",
		"?name=test",
		"=.proplist=.id,name,target,comment,max-limit",
	}

	//when
	device := NewDevice(client, "foo")
	device.Print(entity)

	//then
	if len(client.Sentece) != len(sentences) {
		t.Fatalf("Invalid number of elements in sentence. Expected: %d, got: %d\n", len(sentences), len(client.Sentece))
	}

	for i, v := range sentences {
		if client.Sentece[i] != v {
			t.Fatalf("Incorrect sentence. Expected %s, got: %s\n", v, client.Sentece[i])
		}
	}
}

func TestDeviceRemove(t *testing.T) {
	//given
	client := &ClientMock{}
	entity := Queue{
		ID: "1",
	}

	sentences := []string{
		"/queue/simple/remove",
		"=.id=1",
	}

	//when
	device := NewDevice(client, "foo")
	device.Remove(entity)

	//then
	if len(client.Sentece) != len(sentences) {
		t.Fatalf("Invalid number of elements in sentence. Expected: %d, got: %d\n", len(sentences), len(client.Sentece))
	}

	for i, v := range sentences {
		if client.Sentece[i] != v {
			t.Fatalf("Incorrect sentence. Expected %s, got: %s\n", v, client.Sentece[i])
		}
	}
}

func TestDeviceRemoveNoId(t *testing.T) {
	//given
	client := &ClientMock{}
	entity := Queue{}

	//when
	device := NewDevice(client, "foo")
	err := device.Remove(entity)

	//then
	if err == nil {
		t.Fatal("Remove should return error.")
	}

	if err.Error() != "Id is empty.\n" {
		t.Fatalf("Expected error with message: 'Id is empty', got: %s\n", err)
	}
}

func TestDeviceEdit(t *testing.T) {
	//given
	client := &ClientMock{}
	entity := Queue{
		ID:      "1",
		Name:    "test",
		Target:  "1.2.3.4/5",
		Comment: "foobar",
		Limits:  "123/456",
	}

	sentences := []string{
		"/queue/simple/set",
		"=.id=1",
		"=name=test",
		"=target=1.2.3.4/5",
		"=comment=foobar",
		"=max-limit=123/456",
	}

	//when
	device := NewDevice(client, "foo")
	device.Edit(entity)

	//then
	if len(client.Sentece) != len(sentences) {
		t.Fatalf("Invalid number of elements in sentence. Expected: %d, got: %d\n", len(sentences), len(client.Sentece))
	}

	for i, v := range sentences {
		if client.Sentece[i] != v {
			t.Fatalf("Incorrect sentence. Expected %s, got: %s\n", v, client.Sentece[i])
		}
	}
}

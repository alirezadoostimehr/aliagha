package e2e

import (
	"fmt"
	"github.com/dghubble/sling"
	"os/exec"
	"testing"
)

func TestStartingApp(t *testing.T) {
	cmd := exec.Command("../run.sh")
	err := cmd.Start()
	if err != nil {
		t.Fatal(err.Error() + "!")
	}
	err = cmd.Wait()
	if err != nil {
		t.Fatal(err.Error() + "!")
	}
	t.Log(cmd.Output())
	req, err := sling.New().Base("http://127.0.0.1:3030/").Path("flights?departure_city=Athen&arrival_city=London&date=2020-11-04").Request()
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Println(req.Body)
	t.Log(req)
}

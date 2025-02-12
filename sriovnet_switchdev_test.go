package sriovnet

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	utilfs "github.com/Mellanox/sriovnet/pkg/utils/filesystem"
)

type repContext struct {
	Name         string // create files /sys/bus/pci/devices/<vf addr>/physfn/net/<Name> , /sys/class/net/<Name>
	PhysPortName string // conditionally create if string is empty under /sys/class/net/<Name>/phys_port_name
	PhysSwitchID string // conditionally create if string is empty under /sys/class/net/<Name>/phys_switch_id
}

func setUpRepresentorLayout(vfPciAddress string, rep *repContext) error {
	path := filepath.Join(PciSysDir, vfPciAddress, "physfn/net", rep.Name)
	err := utilfs.Fs.MkdirAll(path, os.FileMode(0755))
	if err != nil {
		return err
	}

	path = filepath.Join(NetSysDir, rep.Name)
	err = utilfs.Fs.MkdirAll(path, os.FileMode(0755))
	if err != nil {
		return err
	}

	if rep.PhysPortName != "" {
		physPortNamePath := filepath.Join(NetSysDir, rep.Name, netdevPhysPortName)
		physPortNameFile, _ := utilfs.Fs.Create(physPortNamePath)
		_, err = physPortNameFile.Write([]byte(rep.PhysPortName))
		if err != nil {
			return err
		}
	}

	if rep.PhysSwitchID != "" {
		physSwitchIDPath := filepath.Join(NetSysDir, rep.Name, netdevPhysSwitchID)
		physSwitchIDFile, _ := utilfs.Fs.Create(physSwitchIDPath)
		_, err = physSwitchIDFile.Write([]byte(rep.PhysSwitchID))
		if err != nil {
			return err
		}
	}

	return nil
}

//nolint:unparam
func setupUplinkRepresentorEnv(t *testing.T, uplink *repContext, vfPciAddress string, vfReps []*repContext) func() {
	var err error
	utilfs.Fs = utilfs.NewFakeFs()
	defer func() {
		if err != nil {
			t.Errorf("setupUplinkRepresentorEnv, got %v", err)
		}
	}()

	err = setUpRepresentorLayout(vfPciAddress, uplink)
	for _, rep := range vfReps {
		err = setUpRepresentorLayout(vfPciAddress, rep)
	}

	return func() { utilfs.Fs.RemoveAll("/") } //nolint:errcheck
}

func TestGetUplinkRepresentorWithPhysPortNameSuccess(t *testing.T) {
	vfPciAddress := "0000:03:00.4"
	uplinkRep := &repContext{"eth0", "p0", "111111"}
	vfsReps := []*repContext{{"enp_0", "pf0vf0", "0123"},
		{"enp_1", "pf0vf1", "0124"}}

	teardown := setupUplinkRepresentorEnv(t, uplinkRep, vfPciAddress, vfsReps)
	defer teardown()
	uplinkNetdev, err := GetUplinkRepresentor(vfPciAddress)
	assert.NoError(t, err)
	assert.Equal(t, "eth0", uplinkNetdev)
}

func TestGetUplinkRepresentorWithoutPhysPortNameSuccess(t *testing.T) {
	vfPciAddress := "0000:03:00.4"
	uplinkRep := &repContext{Name: "eth0", PhysSwitchID: "111111"}
	var vfsReps []*repContext

	teardown := setupUplinkRepresentorEnv(t, uplinkRep, vfPciAddress, vfsReps)
	defer teardown()
	uplinkNetdev, err := GetUplinkRepresentor(vfPciAddress)
	assert.NoError(t, err)
	assert.Equal(t, "eth0", uplinkNetdev)
}

func TestGetUplinkRepresentorWithPhysPortNameFailed(t *testing.T) {
	vfPciAddress := "0000:03:00.4"
	uplinkRep := &repContext{"eth0", "invalid", "111111"}
	vfsReps := []*repContext{{"enp_0", "pf0vf0", "0123"},
		{"enp_1", "pf0vf1", "0124"}}

	expectedError := fmt.Sprintf("uplink for %s not found", vfPciAddress)
	teardown := setupUplinkRepresentorEnv(t, uplinkRep, vfPciAddress, vfsReps)
	defer teardown()
	uplinkNetdev, err := GetUplinkRepresentor(vfPciAddress)
	assert.Error(t, err)
	assert.Equal(t, "", uplinkNetdev)
	assert.Equal(t, expectedError, err.Error())
}

func TestGetUplinkRepresentorErrorMissingSwID(t *testing.T) {
	vfPciAddress := "0000:03:00.4"
	uplinkRep := &repContext{Name: "eth0", PhysPortName: "p0"}
	vfsReps := []*repContext{{Name: "enp_0", PhysPortName: "pf0vf0"},
		{Name: "enp_1", PhysPortName: "pf0vf1"}}
	expectedError := fmt.Sprintf("uplink for %s not found", vfPciAddress)
	teardown := setupUplinkRepresentorEnv(t, uplinkRep, vfPciAddress, vfsReps)
	defer teardown()
	uplinkNetdev, err := GetUplinkRepresentor(vfPciAddress)
	assert.Error(t, err)
	assert.Equal(t, "", uplinkNetdev)
	assert.Equal(t, expectedError, err.Error())
}

func TestGetUplinkRepresentorErrorEmptySwID(t *testing.T) {
	var testErr error
	vfPciAddress := "0000:03:00.4"
	uplinkRep := &repContext{"eth0", "", ""}
	var vfsReps []*repContext
	expectedError := fmt.Sprintf("uplink for %s not found", vfPciAddress)
	teardown := setupUplinkRepresentorEnv(t, uplinkRep, vfPciAddress, vfsReps)
	defer teardown()
	swIDFile := filepath.Join(NetSysDir, "eth0", netdevPhysSwitchID)
	swID, testErr := utilfs.Fs.Create(swIDFile)
	defer func() {
		if testErr != nil {
			t.Errorf("setupUplinkRepresentorEnv, got %v", testErr)
		}
	}()
	_, testErr = swID.Write([]byte(""))
	uplinkNetdev, err := GetUplinkRepresentor(vfPciAddress)
	assert.Error(t, err)
	assert.Equal(t, "", uplinkNetdev)
	assert.Equal(t, expectedError, err.Error())
}

func TestGetUplinkRepresentorErrorMissingUplink(t *testing.T) {
	vfPciAddress := "0000:03:00.4"
	expectedError := fmt.Sprintf("failed to lookup %s", vfPciAddress)
	uplinkNetdev, err := GetUplinkRepresentor(vfPciAddress)
	assert.Error(t, err)
	assert.Equal(t, "", uplinkNetdev)
	assert.Contains(t, err.Error(), expectedError)
}

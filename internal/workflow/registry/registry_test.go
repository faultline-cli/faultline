package registry

import "testing"

func TestDefaultRegistryIncludesCoreSteps(t *testing.T) {
	reg := Default()
	for _, name := range []string{"which_command", "file_exists", "detect_package_manager", "install_package", "noop", "fail"} {
		if _, ok := reg.Lookup(name); !ok {
			t.Fatalf("expected step %s in registry", name)
		}
	}
}

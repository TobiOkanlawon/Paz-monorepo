package web_app

import (
	"log"
	"os"
	"testing"
	"testing/fstest"
)

func setupSuite(t *testing.T) func(t *testing.T) {
	log.Println("setting up a clean partialsManager")
	ClearPartialsManager()
	return func(t *testing.T) {
		log.Println("teardown suite")
	}
}

func TestPartials(t *testing.T) {
	t.Run("can successfully register a partial", func(t *testing.T) {
		teardownSuite := setupSuite(t)
		defer teardownSuite(t)
		path := "test"
		fsMap := fstest.MapFS{
			path: {Data: []byte("test")},
		}
		partials_manager := GetPartialsManager(fsMap)
		err := partials_manager.RegisterPartial("testPartial", path)

		if err != nil {
			t.Fatalf("should not return an error while trying to register a partial %q", err)
		}
	})

	t.Run("returns an error on non-existent path", func(t *testing.T) {
		teardownSuite := setupSuite(t)
		defer teardownSuite(t)
		fsMap := fstest.MapFS{}
		partials_manager := GetPartialsManager(fsMap)
		path := "non-existent-path"
		err := partials_manager.RegisterPartial("testPartial", path)

		if err == nil {
			t.Fatal("should return an error while trying to register a partial at a non-existent path")
		}
	})

	t.Run("can retrieve a partial that was previously stored", func(t *testing.T) {
		teardownSuite := setupSuite(t)
		defer teardownSuite(t)
		path := "testPartial"
		fsMap := fstest.MapFS{
			path: {Data: []byte("some data")},
		}

		partials_manager := GetPartialsManager(fsMap)
		partials_manager.RegisterPartial("test", path)

		returnedPath, err := partials_manager.GetPartial("test")
		if err != nil {
			t.Errorf("did not expect an error: {%q} while trying to retrieve a registered partial: {%q}", err, "test")
		}

		if returnedPath != path {
			t.Errorf("did not return the expected path %q, returned %q", path, returnedPath)
		}
	})

	t.Run("returns an error while trying to retrieve a non-existent partial", func(t *testing.T) {
		teardownSuite := setupSuite(t)
		defer teardownSuite(t)
		manager := GetPartialsManager(fstest.MapFS{})

		_, err := manager.GetPartial("some-non-existent-partial")

		if err == nil {
			t.Error("expected an error while trying to retrieve a non-existent-partial from the partial store")
		}
	})

	t.Run("GetPartialsManager returns a singleton object", func(t *testing.T) {
		teardownSuite := setupSuite(t)
		defer teardownSuite(t)
		path := "non-existent-path"
		fsMap := fstest.MapFS{
			path: {
				Data: []byte("test"),
			},
		}
		partials_manager := GetPartialsManager(fsMap)

		_ = partials_manager.RegisterPartial("testPartial", path)

		_, _ = partials_manager.GetPartial("testPartial")
		partials_manager = GetPartialsManager(os.DirFS("/tmp/fake-folder/"))

		_, err := partials_manager.GetPartial("testPartial")

		if err != nil {
			t.Errorf("should return a singleton, which retains previous data")
		}
	})
}
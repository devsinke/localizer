package localizer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devsinke/localizer"
)

func TestLocalizerBasics(t *testing.T) {
	tempDir := t.TempDir()

	enDir := filepath.Join(tempDir, "en")
	if err := os.MkdirAll(enDir, 0755); err != nil {
		t.Fatal(err)
	}

	enYaml := `
welcome: "Welcome {user}!"
balance_error: "Insufficient funds for account {{.account}}."
`
	if err := os.WriteFile(filepath.Join(enDir, "response.yaml"), []byte(enYaml), 0644); err != nil {
		t.Fatal(err)
	}

	loc := localizer.New(tempDir)
	loc.SetupLocalization()

	msg := loc.Translate("response.welcome", map[string]string{"user": "Alex"})
	if msg != "Welcome Alex!" {
		t.Fatalf("expected 'Welcome Alex!', got '%s'", msg)
	}

	msg2 := loc.Translate("response.balance_error", map[string]string{"account": "ACC100"})
	if msg2 != "Insufficient funds for account ACC100." {
		t.Fatalf("expected 'Insufficient funds for account ACC100.', got '%s'", msg2)
	}
}

func TestLocalizerWithWalletServiceRealAssets(t *testing.T) {
	realAssetDir := "/home/kiprono-alex/Projects/spotsbook/wallet-service/localizations_src"
	if _, err := os.Stat(realAssetDir); os.IsNotExist(err) {
		t.Skip("real assets not present, skipping")
	}

	loc := localizer.New(realAssetDir)
	loc.SetDefaultLocale("en", "en")
	loc.SetupLocalization()

	res := loc.Translate("response.mpesa_deposit_initiated")
	t.Logf("Translated result from wallet-service assets: %s", res)
	if res == "response.mpesa_deposit_initiated" {
		t.Fatalf("Key was not resolved from real wallet-service response.yaml")
	}
}

func TestLocalizerThreadSafety(t *testing.T) {
	tempDir := t.TempDir()

	for _, lang := range []string{"en", "fr", "sw"} {
		dir := filepath.Join(tempDir, lang)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		yamlContent := "hello: \"Hello in " + lang + "\"\n"
		if err := os.WriteFile(filepath.Join(dir, "msg.yaml"), []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}
	}

	loc := localizer.New(tempDir)
	loc.SetupLocalization()

	done := make(chan bool)
	for i := 0; i < 20; i++ {
		go func(idx int) {
			lang := "en"
			if idx%3 == 1 {
				lang = "fr"
			} else if idx%3 == 2 {
				lang = "sw"
			}
			msg := loc.T(lang, "msg.hello")
			if msg != "Hello in "+lang {
				t.Errorf("expected 'Hello in %s', got '%s'", lang, msg)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 20; i++ {
		<-done
	}
}

package config

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestCustomSoundLifecycle(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	src := filepath.Join(t.TempDir(), "yippee.MP3")
	if err := os.WriteFile(src, []byte("not really audio"), 0o644); err != nil {
		t.Fatal(err)
	}

	name, err := ImportSound(src)
	if err != nil {
		t.Fatalf("ImportSound: %v", err)
	}
	if name != "yippee.MP3" {
		t.Fatalf("name = %q, want yippee.MP3", name)
	}

	want := filepath.Join(home, ".config", "yubi-ki-yay", "sounds", "yippee.MP3")
	if got := SoundPath(name); got != want {
		t.Fatalf("SoundPath = %q, want %q", got, want)
	}
	if got := SoundPath("Glass"); got != "/System/Library/Sounds/Glass.aiff" {
		t.Fatalf("SoundPath(Glass) = %q", got)
	}

	custom, err := CustomSounds()
	if err != nil || !slices.Equal(custom, []string{"yippee.MP3"}) {
		t.Fatalf("CustomSounds = %v, %v", custom, err)
	}

	if err := RemoveSound("Glass"); err == nil {
		t.Fatal("RemoveSound(Glass) succeeded, want error")
	}
	if err := RemoveSound(name); err != nil {
		t.Fatalf("RemoveSound: %v", err)
	}
	if custom, _ := CustomSounds(); len(custom) != 0 {
		t.Fatalf("CustomSounds after remove = %v", custom)
	}
}

func TestImportSoundRejectsNonAudio(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	src := filepath.Join(t.TempDir(), "notes.txt")
	if err := os.WriteFile(src, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportSound(src); err == nil {
		t.Fatal("ImportSound(.txt) succeeded, want error")
	}
}

func TestSoundPathStaysInSoundsDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	want := filepath.Join(home, ".config", "yubi-ki-yay", "sounds", "passwd.wav")
	if got := SoundPath("../../../../etc/passwd.wav"); got != want {
		t.Fatalf("SoundPath = %q, want %q", got, want)
	}
}

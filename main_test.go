package main

import (
    "testing"
)

func TestConfGet_Miss(t *testing.T) {
    if confGet("not a key") != "" {
        t.Error("Should return empty string")
    }
}

func TestConfGetInt_Miss(t *testing.T) {
    if confGetInt("not a key") != 0 {
        t.Error("Should return 0")
    }
}

func TestConfGet(t *testing.T) {
    k, v := "some-key", "some-val"
    parsedConfig[k] = v
    x := confGet(k)
    if x != v {
        t.Error("Should return", v, "instead of", x)
    }
}

func TestConfGetInt(t *testing.T) {
    k, v := "some-key", 14
    parsedConfig[k] = float64(v)
    x := confGetInt(k)
    if x != v {
        t.Error("Should return", v, "instead of", x)
    }
}


package auth

import (
	"fmt"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestNewProviderByName(t *testing.T) {
	// Invalid
	provider, err := NewProviderByName("invalid")

	assert.Equal(t, fmt.Errorf("missing provider invalid"), err)
	assert.Equal(t, nil, provider)

	// Apple
	provider, err = NewProviderByName(NameApple)

	assert.Equal(t, nil, err)
	if _, ok := provider.(*Apple); !ok {
		t.Errorf("expected provider to by of type *Apple")
	}

	// Google
	provider, err = NewProviderByName(NameGoogle)

	assert.Equal(t, nil, err)
	if _, ok := provider.(*Google); !ok {
		t.Errorf("expected provider to by of type *Google")
	}

	// Facebook
	provider, err = NewProviderByName(NameFacebook)

	assert.Equal(t, nil, err)
	if _, ok := provider.(*Facebook); !ok {
		t.Errorf("expected provider to by of type *Facebook")
	}
}

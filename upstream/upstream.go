// Package upstream upstream backends
package upstream

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Akagi201/udplb/cache"
	"github.com/Akagi201/udplb/config"
	"github.com/Akagi201/udplb/log"
)

const (
	HT_REMOTE_IP = iota + 1
	HT_REMOTE_PORT
	HT_NONE
)

const (
	TT_STATIC = iota + 1
)

type IDynamicUpstreamSource interface {
	UpdatePeers() ([]string, error)
}

type Upstream struct {
	Config     *config.Upstream
	Targets    []string
	TargetType uint8
	RRcounter  uint
	IsDynamic  bool

	HashType  uint8
	HashCache *cache.CacheManager

	DynamicSource     IDynamicUpstreamSource
	DynamicSourceLock sync.Mutex
}

// GetRRPeer will get a peer from the peers list with a round-robin behavior
func (m *Upstream) GetRRPeer() (*string, error) {
	numTargets := len(m.Targets)

	if numTargets == 0 {
		return nil, errors.New("No target present in upstream")
	}

	peer := &m.Targets[m.RRcounter%uint(numTargets)]

	// in case the number of targets got reduced
	m.RRcounter++

	return peer, nil
}

// GetPeer will fetch a peer either from the cache or the peers list upon request
func (m *Upstream) GetPeer(hash uint32) (*string, error) {
	if m.HashType == HT_NONE {
		return m.GetRRPeer()
	}

	m.DynamicSourceLock.Lock()
	ipeer := m.HashCache.Get(hash)
	if ipeer == nil {
		peer, err := m.GetRRPeer()
		if err != nil {
			return nil, err
		}

		m.HashCache.Add(hash, peer)
		ipeer = peer
	}
	m.DynamicSourceLock.Unlock()

	return ipeer.(*string), nil
}

// UpdateDynamicPeers will update the peers of a dynamic upstream
func (m *Upstream) UpdateDynamicPeers() {
	if m.IsDynamic {
		targets, err := m.DynamicSource.UpdatePeers()
		if err != nil {
			log.Errorf("Can't update upstream (%s): %s", m.Config.Name, err)
			return
		}

		m.DynamicSourceLock.Lock()
		m.Targets = targets

		// Check if an old peer entry is stored in the cache, if yes, delete it
		storedKeys := m.HashCache.GetKeys()
		for _, sk := range storedKeys {
			found := false
			for _, t := range targets {
				if sk == binary.BigEndian.Uint32(net.ParseIP(t)) {
					found = true
				}
			}

			if !found {
				m.HashCache.DeleteEntry(sk)
			}
		}
		m.DynamicSourceLock.Unlock()
	} else {
		fmt.Println()
	}
}

// AutoUpdatePeer is a helper function that will update the list of dynamic upstream peers at a regular rate defined by
// the parameter seconds.
func AutoUpdatePeer(us *Upstream, seconds int) {
	ticker := time.NewTicker(time.Second * time.Duration(seconds))

	us.UpdateDynamicPeers()
	for range ticker.C {
		us.UpdateDynamicPeers()
	}
}

// NewUpstream parses an upstream configuration block and creates an Upstream object
func NewUpstream(cfg *config.Upstream) (*Upstream, error) {
	var hashType uint8

	switch cfg.Hash {
	case "remote_ip":
		hashType = HT_REMOTE_IP
		break

	case "remote_port":
		hashType = HT_REMOTE_PORT
		break

	default:
		log.Warn("Incorrect upstream hash. Defaulting to none.")

	case "":
	case "none":
		hashType = HT_NONE
		break
	}

	us := &Upstream{
		Config:    cfg,
		RRcounter: 0,
		HashType:  hashType,
		IsDynamic: false,
	}

	if us.HashType != HT_NONE {
		us.HashCache = cache.MustNewCacheManager(60, 5)
	}

	switch cfg.Type {
	case "":
	case "static":
		us.TargetType = TT_STATIC
		us.Targets = cfg.Targets
	default:
		log.Warn("Incorrect upstream type. Defaulting to none.")
	}

	return us, nil
}

// MustNewUpstream does the same thing as NewUpstream and will panic if the creation fails
func MustNewUpstream(cfg *config.Upstream) *Upstream {
	us, err := NewUpstream(cfg)
	if err != nil {
		panic(err)
	}

	return us
}

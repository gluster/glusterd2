package api

import (
	"encoding/xml"
)

// FileGfID represents the file details on a volume
type FileGfID struct {
	Filename string `xml:",chardata"`
	GfID     string `xml:"gfid,attr"`
}

// BrickHealInfo represents brick details for Heal Info command
type BrickHealInfo struct {
	HostID                    string     `xml:"hostUuid,attr" json:"host-id"`
	Name                      string     `xml:"name" json:"name"`
	Status                    string     `xml:"status" json:"status"`
	TotalEntriesRaw           string     `xml:"totalNumberOfEntries" json:"-"`
	EntriesInHealPendingRaw   string     `xml:"numberOfEntriesInHealPending" json:"-"`
	EntriesInSplitBrainRaw    string     `xml:"numberOfEntriesInSplitBrain" json:"-"`
	EntriesPossiblyHealingRaw string     `xml:"numberOfEntriesPossiblyHealing" json:"-"`
	EntriesRaw                string     `xml:"numberOfEntries" json:"-"`
	TotalEntries              *int64     `json:"total-entries,omitempty"`
	EntriesInHealPending      *int64     `json:"entries-in-heal-pending,omitempty"`
	EntriesInSplitBrain       *int64     `json:"entries-in-split-brain,omitempty"`
	EntriesPossiblyHealing    *int64     `json:"entries-possibly-healing,omitempty"`
	Entries                   *int64     `json:"entries,omitempty"`
	Files                     []FileGfID `xml:"file" json:"file-gfid,omitempty"`
}

// HealInfo represents structure of stdout while running glfsheal binary
type HealInfo struct {
	XMLNAME xml.Name        `xml:"cliOutput"`
	Bricks  []BrickHealInfo `xml:"healInfo>bricks>brick"`
}

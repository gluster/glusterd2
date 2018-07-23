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
	HostID                 string     `xml:"hostUuid,attr" json:"host-id"`
	Name                   string     `xml:"name" json:"name"`
	Status                 string     `xml:"status" json:"status"`
	TotalEntries           *int       `xml:"totalNumberOfEntries" json:"total-entries,omitempty"`
	EntriesInHealPending   *int       `xml:"numberOfEntriesInHealPending" json:"entries-in-heal-pending,omitempty"`
	EntriesInSplitBrain    *int       `xml:"numberOfEntriesInSplitBrain" json:"entries-in-split-brain,omitempty"`
	EntriesPossiblyHealing *int       `xml:"numberOfEntriesPossiblyHealing" json:"entries-possibly-healing,omitempty"`
	Entries                *int       `xml:"numberOfEntries" json:"entries,omitempty"`
	Files                  []FileGfID `xml:"file" json:"file-gfid,omitempty"`
}

// HealInfo represents structure of stdout while running glfsheal binary
type HealInfo struct {
	XMLNAME xml.Name        `xml:"cliOutput"`
	Bricks  []BrickHealInfo `xml:"healInfo>bricks>brick"`
}

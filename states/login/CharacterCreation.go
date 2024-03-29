package login

import (
	"fmt"

	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/states/game"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-server/data"
	"github.com/chimera-rpg/go-server/network"
)

// CharacterCreation is our State for connecting as, creating, or deleting a
// character.
type CharacterCreation struct {
	client.State
	layout       ui.LayoutEntry
	bail         chan bool
	selectChan   chan Selection
	tabChan      chan string
	selection    Selection
	genusEntries []*entry
	focusedTab   ui.ElementI
	// ugh
	attributeTexts map[ui.ElementI][]ui.ElementI
}

type Selection struct {
	genus    string
	species  string
	variety  string
	culture  string
	training string
}

type entrySelection struct {
	container *ui.Container
	name      ui.ElementI
	image     ui.ElementI
	selected  bool
}

type entryInfo struct {
	container   *ui.Container
	description ui.ElementI
}

type entry struct {
	animID     uint32
	faceID     uint32
	name       string
	received   bool
	selection  entrySelection
	info       entryInfo
	children   []*entry
	attributes data.AttributeSets
}

func (s *CharacterCreation) makeEntrySelection(name string, imageID uint32, attributes data.AttributeSets, selection Selection) entrySelection {
	var err error
	var container *ui.Container
	container, err = ui.NewContainerElement(ui.ContainerConfig{
		Style: s.Client.DataManager.Styles["Creation"]["EntrySelection"],
		Events: ui.Events{
			OnPressed: func(button uint8, x, y int32) bool {
				s.selectChan <- selection
				return true
			},
		},
	})
	if err != nil {
		panic(err)
	}

	nameEl := ui.NewTextElement(ui.TextElementConfig{
		Style: s.Client.DataManager.Styles["Creation"]["EntrySelection__name"],
		Value: name,
	})

	config := ui.ImageElementConfig{
		Style: s.Client.DataManager.Styles["Creation"]["EntrySelection__image"],
	}
	imageID = 0
	if imageID == 0 {
		imageData, err := s.Client.DataManager.GetImage(s.Client.DataManager.GetDataPath("ui/loading.png"))
		if err != nil {
			panic(err)
		}
		config.Image = imageData
	} else {
		config.ImageID = imageID
	}
	imageEl := ui.NewImageElement(config)

	container.GetAdoptChannel() <- nameEl
	container.GetAdoptChannel() <- imageEl

	return entrySelection{
		container: container,
		name:      nameEl,
		image:     imageEl,
	}
}

func (s *CharacterCreation) attributesToStrings(attr data.Attributes) (str []string) {
	if attr.Might < 0 {
		str = append(str, fmt.Sprintf("%d Might", attr.Might))
	} else if attr.Might > 0 {
		str = append(str, fmt.Sprintf("+%d Might", attr.Might))
	}
	if attr.Prowess < 0 {
		str = append(str, fmt.Sprintf("%d Prowess", attr.Prowess))
	} else if attr.Prowess > 0 {
		str = append(str, fmt.Sprintf("+%d Prowess", attr.Prowess))
	}
	if attr.Focus < 0 {
		str = append(str, fmt.Sprintf("%d Focus", attr.Focus))
	} else if attr.Focus > 0 {
		str = append(str, fmt.Sprintf("+%d Focus", attr.Focus))
	}
	if attr.Sense < 0 {
		str = append(str, fmt.Sprintf("%d Sense", attr.Sense))
	} else if attr.Sense > 0 {
		str = append(str, fmt.Sprintf("+%d Sense", attr.Sense))
	}
	if attr.Haste < 0 {
		str = append(str, fmt.Sprintf("%d Haste", attr.Haste))
	} else if attr.Haste > 0 {
		str = append(str, fmt.Sprintf("+%d Haste", attr.Haste))
	}
	if attr.Reaction < 0 {
		str = append(str, fmt.Sprintf("%d Reaction", attr.Reaction))
	} else if attr.Reaction > 0 {
		str = append(str, fmt.Sprintf("+%d Reaction", attr.Reaction))
	}
	return
}

func (s *CharacterCreation) makeEntryInfo(description string, attributes data.AttributeSets) entryInfo {
	container, err := ui.NewContainerElement(ui.ContainerConfig{
		Style: s.Client.DataManager.Styles["Creation"]["EntryInfo"],
		Events: ui.Events{
			OnPressed: func(button uint8, x, y int32) bool {
				// TODO
				return true
			},
		},
	})
	if err != nil {
		panic(err)
	}

	descContainer, _ := ui.NewContainerElement(ui.ContainerConfig{
		Style: s.Client.DataManager.Styles["Creation"]["EntryInfo__description"],
	})

	descEl := ui.NewTextElement(ui.TextElementConfig{
		Style: s.Client.DataManager.Styles["Creation"]["EntryInfo__description__text"],
		Value: description,
	})

	descContainer.GetAdoptChannel() <- descEl

	attrEl, err := ui.NewContainerElement(ui.ContainerConfig{
		Style: s.Client.DataManager.Styles["Creation"]["EntryInfo__attributes"],
	})

	makeSection := func(title string, which string, attr data.Attributes) {
		container, _ := ui.NewContainerElement(ui.ContainerConfig{
			Style: s.Client.DataManager.Styles["Creation"]["EntryInfo__"+which],
		})

		titleEl := ui.NewTextElement(ui.TextElementConfig{
			Style: s.Client.DataManager.Styles["Creation"]["EntryInfo__"+which+"__title"],
			Value: title,
		})

		attrs, _ := ui.NewContainerElement(ui.ContainerConfig{
			Style: s.Client.DataManager.Styles["Creation"]["EntryInfo__"+which+"__attributes"],
		})

		for _, str := range s.attributesToStrings(attr) {
			t := "good"
			if str[0] == '-' {
				t = "bad"
			}
			el := ui.NewTextElement(ui.TextElementConfig{
				Style: s.Client.DataManager.Styles["Creation"]["EntryInfo__attribute__"+t],
				Value: str,
			})
			attrs.GetAdoptChannel() <- el
		}

		container.GetAdoptChannel() <- titleEl
		container.GetAdoptChannel() <- attrs

		attrEl.GetAdoptChannel() <- container
	}

	makeSection("Physical", "physical", attributes.Physical)
	makeSection("Arcane", "arcane", attributes.Arcane)
	makeSection("Spirit", "spirit", attributes.Spirit)

	container.GetAdoptChannel() <- attrEl
	container.GetAdoptChannel() <- descContainer

	return entryInfo{
		container: container,
	}
}

// Init is our CharacterCreation init state.
func (s *CharacterCreation) Init(t interface{}) (next client.StateI, nextArgs interface{}, err error) {
	s.bail = make(chan bool, 1)
	s.selectChan = make(chan Selection, 1)
	s.tabChan = make(chan string, 1)
	s.Client.Log.Print("CharacterCreation State")

	s.layout = s.Client.DataManager.Layouts["Creation"][0].Generate(s.Client.DataManager.Styles["Creation"], map[string]interface{}{
		"Container": ui.ContainerConfig{
			Value: "Creation",
		},
		"Characters": ui.ContainerConfig{
			Value: "Character",
		},
		"CharacterName": ui.InputElementConfig{
			Placeholder: "name",
			Events: ui.Events{
				OnKeyDown: func(char uint8, modifiers uint16, repeat bool) bool {
					/*if char == 13 { // Enter
						s.layout.Find("CreateButton").Element.OnPressed(1, 0, 0)
					}*/
					return true
				},
			},
		},
		"CharacterDescription": ui.InputElementConfig{
			Placeholder: "description",
			Events: ui.Events{
				OnKeyDown: func(char uint8, modifiers uint16, repeat bool) bool {
					/*if char == 13 { // Enter
						s.layout.Find("CreateButton").Element.OnPressed(1, 0, 0)
					}*/
					return true
				},
			},
		},
		"CreateButton": ui.ButtonElementConfig{
			Value: "Create Character",
			Events: ui.Events{
				OnPressed: func(button uint8, x int32, y int32) bool {
					s.Client.Send(network.Command(network.CommandCreateCharacter{
						Name: s.layout.Find("CharacterName").Element.GetValue(),
					}))
					return false
				},
			},
		},
		"BackButton": ui.ButtonElementConfig{
			Value: "Back",
			Events: ui.Events{
				OnPressed: func(button uint8, x int32, y int32) bool {
					s.bail <- true
					s.Client.StateChannel <- client.StateMessage{Pop: true}
					return false
				},
			},
		},
		//
		"Nature__Tab": ui.ButtonElementConfig{
			Value: "Nature",
			Events: ui.Events{
				OnPressed: func(button uint8, x int32, y int32) bool {
					s.tabChan <- "Nature"
					return false
				},
			},
		},
		"Nurture__Tab": ui.ButtonElementConfig{
			Value: "Nurture",
			Events: ui.Events{
				OnPressed: func(button uint8, x int32, y int32) bool {
					s.tabChan <- "Nurture"
					return false
				},
			},
		},
		"Results__Attributes": ui.ContainerConfig{
			Style: s.Client.DataManager.Styles["Creation"]["EntryInfo__attributes"],
		},
		"Results__Attributes__Physical": ui.ContainerConfig{
			Style: s.Client.DataManager.Styles["Creation"]["EntryInfo__physical"],
		},
		"Results__Attributes__Arcane": ui.ContainerConfig{
			Style: s.Client.DataManager.Styles["Creation"]["EntryInfo__arcane"],
		},
		"Results__Attributes__Spirit": ui.ContainerConfig{
			Style: s.Client.DataManager.Styles["Creation"]["EntryInfo__spirit"],
		},
		"Results__Attributes__Physical__Title": ui.TextElementConfig{
			Value: "Physical",
			Style: s.Client.DataManager.Styles["Creation"]["EntryInfo__physical__title"],
		},
		"Results__Attributes__Physical__Attributes": ui.ContainerConfig{
			Style: s.Client.DataManager.Styles["Creation"]["EntryInfo__physical__attributes"],
		},
		"Results__Attributes__Arcane__Title": ui.TextElementConfig{
			Value: "Arcane",
			Style: s.Client.DataManager.Styles["Creation"]["EntryInfo__arcane__title"],
		},
		"Results__Attributes__Arcane__Attributes": ui.ContainerConfig{
			Style: s.Client.DataManager.Styles["Creation"]["EntryInfo__arcane__attributes"],
		},
		"Results__Attributes__Spirit__Title": ui.TextElementConfig{
			Value: "Spirit",
			Style: s.Client.DataManager.Styles["Creation"]["EntryInfo__spirit__title"],
		},
		"Results__Attributes__Spirit__Attributes": ui.ContainerConfig{
			Style: s.Client.DataManager.Styles["Creation"]["EntryInfo__spirit__attributes"],
		},
	})

	s.layout.Find("Nurture__Content").Element.GetUpdateChannel() <- ui.UpdateHidden(true)

	s.attributeTexts = make(map[ui.ElementI][]ui.ElementI)
	// Create results attributes.
	{
		//s.layout.Find("Results__Attributes__Physical").Element.GetUpdateChannel() <- ui.UpdateHidden(true)
	}

	s.Client.RootWindow.AdoptChannel <- s.layout.Find("Container").Element

	// Let the server know we're ready!
	s.Client.Send(network.Command(network.CommandQueryGenera{}))

	go s.Loop()

	return
}

func (s *CharacterCreation) refreshImages() {
	var iter func(entry *entry)
	iter = func(entry *entry) {
		anim := s.Client.DataManager.GetAnimation(entry.animID)
		face := anim.GetFace(entry.faceID)
		if len(face.Frames) > 0 {
			entry.selection.image.GetUpdateChannel() <- ui.UpdateImageID(face.Frames[0].ImageID)
		}
		for _, ch := range entry.children {
			iter(ch)
		}
	}
	for _, entry := range s.genusEntries {
		iter(entry)
	}
}

func (s *CharacterCreation) addGenus(genus network.Genus) {
	anim := s.Client.DataManager.GetAnimation(genus.AnimationID)
	face := anim.GetFace(genus.FaceID)
	imageID := uint32(0)
	if len(face.Frames) > 0 {
		imageID = face.Frames[0].ImageID
	}
	entry := entry{
		name:       genus.Name,
		animID:     genus.AnimationID,
		faceID:     genus.FaceID,
		selection:  s.makeEntrySelection(genus.Name, imageID, genus.Attributes, Selection{genus: genus.Name}),
		info:       s.makeEntryInfo(genus.Description, genus.Attributes),
		attributes: genus.Attributes,
		received:   false,
	}
	s.layout.Find("Genera__List").Element.GetAdoptChannel() <- entry.selection.container
	s.layout.Find("Genera__Info").Element.GetAdoptChannel() <- entry.info.container
	entry.info.container.GetUpdateChannel() <- ui.UpdateHidden(true)
	s.genusEntries = append(s.genusEntries, &entry)
}

func (s *CharacterCreation) addSpecies(genus string, species network.Species) {
	for _, g := range s.genusEntries {
		if g.name != genus {
			continue
		}
		g.received = true
		anim := s.Client.DataManager.GetAnimation(species.AnimationID)
		face := anim.GetFace(species.FaceID)
		imageID := uint32(0)
		if len(face.Frames) > 0 {
			imageID = face.Frames[0].ImageID
		}
		entry := entry{
			name:       species.Name,
			animID:     species.AnimationID,
			faceID:     species.FaceID,
			selection:  s.makeEntrySelection(species.Name, imageID, species.Attributes, Selection{genus: genus, species: species.Name}),
			info:       s.makeEntryInfo(species.Description, species.Attributes),
			attributes: species.Attributes,
			received:   false,
		}
		if s.selection.genus != genus {
			entry.selection.container.SetHidden(true)
		}
		s.layout.Find("Species__List").Element.GetAdoptChannel() <- entry.selection.container
		s.layout.Find("Species__Info").Element.GetAdoptChannel() <- entry.info.container
		entry.info.container.GetUpdateChannel() <- ui.UpdateHidden(true)
		g.children = append(g.children, &entry)
	}
}

func (s *CharacterCreation) addVariety(genus string, species string, variety network.Variety) {
	for _, g := range s.genusEntries {
		if g.name != genus {
			continue
		}
		for _, sp := range g.children {
			if sp.name != species {
				continue
			}
			sp.received = true
			anim := s.Client.DataManager.GetAnimation(variety.AnimationID)
			face := anim.GetFace(variety.FaceID)
			imageID := uint32(0)
			if len(face.Frames) > 0 {
				imageID = face.Frames[0].ImageID
			}
			entry := entry{
				name:       variety.Name,
				animID:     variety.AnimationID,
				faceID:     variety.FaceID,
				selection:  s.makeEntrySelection(variety.Name, imageID, variety.Attributes, Selection{genus: genus, species: species, variety: variety.Name}),
				info:       s.makeEntryInfo(variety.Description, variety.Attributes),
				attributes: variety.Attributes,
			}
			if s.selection.species != species {
				entry.selection.container.SetHidden(true)
			}
			s.layout.Find("Variety__List").Element.GetAdoptChannel() <- entry.selection.container
			s.layout.Find("Variety__Info").Element.GetAdoptChannel() <- entry.info.container
			entry.info.container.GetUpdateChannel() <- ui.UpdateHidden(true)
			sp.children = append(sp.children, &entry)
		}
	}
}

// addCharacter adds a button for the provided character name.
func (s *CharacterCreation) addCharacter(name string) {
	isFocused := false
	if name == s.Client.DataManager.Config.Servers[s.Client.CurrentServer].Character {
		isFocused = true
	}

	elChar := ui.NewButtonElement(ui.ButtonElementConfig{
		Style: s.Client.DataManager.Styles["Selection"]["CharacterEntry"],
		Value: name,
		Events: ui.Events{
			OnPressed: func(button uint8, x int32, y int32) bool {
				s.Client.Log.Printf("Logging in with character %s", name)
				s.Client.DataManager.Config.Servers[s.Client.CurrentServer].Character = name
				s.Client.Send(network.Command(network.CommandSelectCharacter{
					Name: name,
				}))
				return false
			},
		},
	})
	if isFocused {
		elChar.Focus()
	}
	s.layout.Find("Characters").Element.GetAdoptChannel() <- elChar
}

func (s *CharacterCreation) getNatures() (genus *entry, species *entry, variety *entry) {
	for _, g := range s.genusEntries {
		if g.name == s.selection.genus {
			genus = g
			for _, sp := range g.children {
				if sp.name == s.selection.species {
					species = sp
					for _, v := range sp.children {
						if v.name == s.selection.variety {
							variety = v
							return
						}
					}
					return
				}
			}
			return
		}
	}
	return
}

func (s *CharacterCreation) refreshResults() {
	var attributes data.AttributeSets
	physEl := s.layout.Find("Results__Attributes__Physical__Attributes")
	arcaneEl := s.layout.Find("Results__Attributes__Arcane__Attributes")
	spiritEl := s.layout.Find("Results__Attributes__Spirit__Attributes")

	genus, species, variety := s.getNatures()
	if genus != nil {
		attributes.Add(genus.attributes)
	}
	if species != nil {
		attributes.Add(species.attributes)
	}
	if variety != nil {
		attributes.Add(variety.attributes)
	}

	dummyImage, err := s.Client.DataManager.GetImage(s.Client.DataManager.GetDataPath("ui/loading.png"))
	if err != nil {
		panic(err)
	}

	// Update the image
	imageEl := s.layout.Find("CharacterImage")
	if variety != nil {
		anim := s.Client.DataManager.GetAnimation(variety.animID)
		face := anim.GetFace(variety.faceID)
		if len(face.Frames) > 0 {
			imageEl.Element.GetUpdateChannel() <- ui.UpdateImageID(face.Frames[0].ImageID)
		} else {
			imageEl.Element.GetUpdateChannel() <- dummyImage
		}
	} else if species != nil {
		anim := s.Client.DataManager.GetAnimation(species.animID)
		face := anim.GetFace(species.faceID)
		if len(face.Frames) > 0 {
			imageEl.Element.GetUpdateChannel() <- ui.UpdateImageID(face.Frames[0].ImageID)
		} else {
			imageEl.Element.GetUpdateChannel() <- dummyImage
		}
	} else if genus != nil {
		anim := s.Client.DataManager.GetAnimation(genus.animID)
		face := anim.GetFace(genus.faceID)
		if len(face.Frames) > 0 {
			imageEl.Element.GetUpdateChannel() <- ui.UpdateImageID(face.Frames[0].ImageID)
		} else {
			imageEl.Element.GetUpdateChannel() <- dummyImage
		}
	} else {
		imageEl.Element.GetUpdateChannel() <- dummyImage
	}

	// FIXME: We should re-use text elements...
	addAttributeStrings := func(attrs ui.ElementI, attr data.Attributes) {
		attrStrings := s.attributesToStrings(attr)
		for i, str := range attrStrings {
			t := "good"
			if str[0] == '-' {
				t = "bad"
			}

			if i < len(s.attributeTexts[attrs]) {
				s.attributeTexts[attrs][i].GetUpdateChannel() <- ui.UpdateParseStyle(s.Client.DataManager.Styles["Creation"]["EntryInfo__attribute__"+t])
				s.attributeTexts[attrs][i].GetUpdateChannel() <- ui.UpdateValue{Value: str}
			} else {
				el := ui.NewTextElement(ui.TextElementConfig{
					Style: s.Client.DataManager.Styles["Creation"]["EntryInfo__attribute__"+t],
					Value: str,
				})
				attrs.GetAdoptChannel() <- el

				s.attributeTexts[attrs] = append(s.attributeTexts[attrs], el)
			}
		}

		// Remove excess.
		for i := len(s.attributeTexts[attrs]) - 1; i >= len(attrStrings); i-- {
			attrs.GetDisownChannel() <- s.attributeTexts[attrs][i]
		}
		s.attributeTexts[attrs] = s.attributeTexts[attrs][:len(attrStrings)]
	}
	addAttributeStrings(physEl.Element, attributes.Physical)
	addAttributeStrings(arcaneEl.Element, attributes.Arcane)
	addAttributeStrings(spiritEl.Element, attributes.Spirit)
}

// Close our CharacterCreation State.
func (s *CharacterCreation) Close() {
	s.layout.Find("Container").Element.GetDestroyChannel() <- true
}

func (s *CharacterCreation) Leave() {
	s.layout.Find("Container").Element.GetUpdateChannel() <- ui.UpdateHidden(true)
}

func (s *CharacterCreation) Enter(args ...interface{}) {
	s.layout.Find("Container").Element.GetUpdateChannel() <- ui.UpdateHidden(false)
}

func (s *CharacterCreation) getEntry(entries []*entry, section string) *entry {
	for _, entry := range entries {
		if entry.name != section {
			continue
		}
		return entry
	}
	return nil
}

func (s *CharacterCreation) getGenus(genus string) *entry {
	for _, g := range s.genusEntries {
		if g.name == genus {
			return g
		}
	}
	return nil
}

// Select selects a set of genera, species, culture, and training.
func (s *CharacterCreation) Select(selection Selection) {
	// Requestasaurus rex the network stuff
	if genus := s.getGenus(selection.genus); genus != nil {
		if !genus.received {
			s.Client.Send(network.Command(network.CommandQuerySpecies{Genus: selection.genus}))
		} else {
			if species := s.getEntry(genus.children, selection.species); species != nil {
				if !species.received {
					s.Client.Send(network.Command(network.CommandQueryVariety{Genus: selection.genus, Species: selection.species}))
				}
			}
		}
	}

	// * Step 1. Hide current genus/species/variety and send network requests as needed.
	// Reset the selection.
	if s.selection.genus != selection.genus {
		if genus := s.getGenus(s.selection.genus); genus != nil {
			// Hide genus info
			genus.info.container.UpdateChannel <- ui.UpdateHidden(true)
			genus.selection.container.GetUpdateChannel() <- ui.UpdateParseStyle(s.Client.DataManager.Styles["Creation"]["EntrySelection"])
			// Hide species.
			for _, entry := range genus.children {
				entry.info.container.UpdateChannel <- ui.UpdateHidden(true)
				entry.selection.container.UpdateChannel <- ui.UpdateHidden(true)
				entry.selection.container.GetUpdateChannel() <- ui.UpdateParseStyle(s.Client.DataManager.Styles["Creation"]["EntrySelection"])
				// Hide varieties.
				for _, entry := range entry.children {
					entry.info.container.UpdateChannel <- ui.UpdateHidden(true)
					entry.selection.container.UpdateChannel <- ui.UpdateHidden(true)
					entry.selection.container.GetUpdateChannel() <- ui.UpdateParseStyle(s.Client.DataManager.Styles["Creation"]["EntrySelection"])
				}
			}
		}
	} else if s.selection.species != selection.species {
		if genus := s.getGenus(s.selection.genus); genus != nil {
			if species := s.getEntry(genus.children, s.selection.species); species != nil {
				// Hide species info
				species.info.container.UpdateChannel <- ui.UpdateHidden(true)
				species.selection.container.GetUpdateChannel() <- ui.UpdateParseStyle(s.Client.DataManager.Styles["Creation"]["EntrySelection"])
				// Hide varieties.
				for _, entry := range species.children {
					entry.info.container.UpdateChannel <- ui.UpdateHidden(true)
					entry.selection.container.UpdateChannel <- ui.UpdateHidden(true)
					entry.selection.container.GetUpdateChannel() <- ui.UpdateParseStyle(s.Client.DataManager.Styles["Creation"]["EntrySelection"])
				}
			}
		}
	} else if s.selection.variety != selection.variety {
		if genus := s.getGenus(s.selection.genus); genus != nil {
			if species := s.getEntry(genus.children, s.selection.species); species != nil {
				if variety := s.getEntry(species.children, s.selection.variety); variety != nil {
					variety.info.container.UpdateChannel <- ui.UpdateHidden(true)
					variety.selection.container.GetUpdateChannel() <- ui.UpdateParseStyle(s.Client.DataManager.Styles["Creation"]["EntrySelection"])
				}
			}
		}
	}

	s.selection = selection

	// * Step 2. Show newly selected genus/species/variety fields.
	if genus := s.getGenus(s.selection.genus); genus != nil {
		genus.info.container.UpdateChannel <- ui.UpdateHidden(false)
		genus.selection.container.GetUpdateChannel() <- ui.UpdateParseStyle(s.Client.DataManager.Styles["Creation"]["EntrySelection--selected"])
		// Show species.
		for _, entry := range genus.children {
			entry.selection.container.UpdateChannel <- ui.UpdateHidden(false)
			if entry.name == s.selection.species {
				entry.info.container.UpdateChannel <- ui.UpdateHidden(false)
				entry.selection.container.GetUpdateChannel() <- ui.UpdateParseStyle(s.Client.DataManager.Styles["Creation"]["EntrySelection--selected"])
				// Show varieties.
				for _, entry := range entry.children {
					entry.selection.container.UpdateChannel <- ui.UpdateHidden(false)
					if entry.name == s.selection.variety {
						entry.info.container.UpdateChannel <- ui.UpdateHidden(false)
						entry.selection.container.GetUpdateChannel() <- ui.UpdateParseStyle(s.Client.DataManager.Styles["Creation"]["EntrySelection--selected"])
					}
				}
			}
		}
	}
	s.refreshResults()
}

func (s *CharacterCreation) Tab(t string) {
	if t == "Nature" {
		s.layout.Find("Nature__Content").Element.GetUpdateChannel() <- ui.UpdateHidden(false)
		s.layout.Find("Nurture__Content").Element.GetUpdateChannel() <- ui.UpdateHidden(true)
	} else if t == "Nurture" {
		s.layout.Find("Nature__Content").Element.GetUpdateChannel() <- ui.UpdateHidden(true)
		s.layout.Find("Nurture__Content").Element.GetUpdateChannel() <- ui.UpdateHidden(false)
	}
}

// Loop is our loop for managing network activity and beyond.
func (s *CharacterCreation) Loop() {
	for {
		if !s.Running {
			continue
		}
		select {
		case <-s.bail:
			return
		case t := <-s.selectChan:
			s.Select(t)
		case t := <-s.tabChan:
			s.Tab(t)
		case cmd := <-s.Client.CmdChan:
			ret := s.HandleNet(cmd)
			if ret {
				return
			}
		case <-s.Client.DataManager.UpdatedImageIDs:
			s.refreshImages()
			// TODO: Refresh genus/species/pc image
		case <-s.Client.ClosedChan:
			s.Client.Log.Print("Lost connection to server.")
			s.Client.StateChannel <- client.StateMessage{PopToTop: true, Args: nil}
			return
		}
	}
}

// HandleNet manages our network communications.
func (s *CharacterCreation) HandleNet(cmd network.Command) bool {
	switch t := cmd.(type) {
	case network.CommandGraphics:
		s.Client.DataManager.HandleGraphicsCommand(t)
		s.refreshImages()
	case network.CommandAnimation:
		s.Client.DataManager.HandleAnimationCommand(t)
		s.refreshImages()
	case network.CommandSound:
		s.Client.DataManager.HandleSoundCommand(t)
	case network.CommandAudio:
		s.Client.DataManager.HandleAudioCommand(t)
	case network.CommandBasic:
		if t.Type == network.Reject {
			s.Client.Log.Printf("Server rejected us: %s\n", t.String)
		} else if t.Type == network.Okay {
			s.Client.Log.Printf("Server accepted us: %s\n", t.String)
			// Might as well save the configuration now.
			if err := s.Client.DataManager.Config.Write(); err != nil {
				s.Client.Log.Errorln(err)
			}
			s.Client.StateChannel <- client.StateMessage{Push: true, State: &game.Game{}, Args: nil}
			return true
		}
	case network.CommandCreateCharacter:
		s.addCharacter(t.Name)
	case network.CommandQueryGenera:
		for _, genus := range t.Genera {
			s.Client.DataManager.EnsureAnimation(genus.AnimationID)
			s.addGenus(genus)
		}
		s.refreshImages()
	case network.CommandQuerySpecies:
		for _, species := range t.Species {
			s.Client.DataManager.EnsureAnimation(species.AnimationID)
			s.addSpecies(t.Genus, species)
		}
		s.refreshImages()
	case network.CommandQueryVariety:
		for _, variety := range t.Variety {
			s.Client.DataManager.EnsureAnimation(variety.AnimationID)
			s.addVariety(t.Genus, t.Species, variety)
		}
		s.refreshImages()
	default:
		s.Client.Log.Printf("Server sent incorrect Command\n")
		s.Client.StateChannel <- client.StateMessage{PopToTop: true, Args: nil}
		return true
	}
	return false
}

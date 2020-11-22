package gui

import "C"
import (
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/cmd/fyne_settings/settings"
	"fyne.io/fyne/container"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	gpg "github.com/Geo25rey/crypto/openpgp"
	"github.com/Geo25rey/crypto/openpgp/packet"
	"github.com/ipfs/go-cid"
	iface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/spf13/viper"
	"io"
	views "ipmail/gui/fyne_views"
	"ipmail/libipmail"
	"ipmail/libipmail/crypto"
	"ipmail/libipmail/util"
	"strings"
	"sync"
)

func hasInvalidCharacters(s string) error {
	for _, c := range s {
		switch c {
		case '(', ')', '<', '>', 0:
			return fmt.Errorf("input contains an invalid character '%c'", c)
		}
	}
	return nil
}

func prompt(window fyne.Window, onResults func([]string, error), isPassword bool, title string, content string, fields ...string) {
	results := make([]string, 0)
	contentLbl := widget.NewLabel(content)
	contentLbl.Wrapping = fyne.TextWrapWord
	formItems := make([]*widget.FormItem, 0)
	for i, field := range fields {
		var input *widget.Entry
		if isPassword {
			input = widget.NewPasswordEntry()
		} else {
			input = widget.NewEntry()
		}
		errorText := widget.NewLabel("")
		input.Validator = hasInvalidCharacters
		input.SetOnValidationChanged(func(err error) {
			if err != nil {
				errorText.SetText(err.Error())
			} else {
				errorText.SetText("")
			}
		})
		input.OnChanged = func(s string) {
			results[i] = s
		}
		inputErrorText := container.NewBorder(nil, nil, nil, errorText, input)
		formItems = append(formItems, widget.NewFormItem(field, inputErrorText))
		results = append(results, "")
	}
	form := &widget.Form{
		Items: formItems,
	}
	dialog.ShowCustomConfirm(title, "Create", "Exit",
		container.NewVBox(widget.NewLabel(title),
			widget.NewSeparator(),
			contentLbl,
			form,
		), func(submitted bool) {
			if submitted {
				go onResults(results, nil)
			} else {
				go onResults(nil, errors.New("user cancelled the form"))
			}
		}, window)
}

func initMenuBar(w *fyne.Window,
	identity *crypto.SelfIdentity, contacts *crypto.ContactsIdentityList,
	sender *ipmail.Sender) {
	a := fyne.CurrentApp()
	shortcutFocused := func(s fyne.Shortcut, w fyne.Window) {
		if focused, ok := w.Canvas().Focused().(fyne.Shortcutable); ok {
			focused.TypedShortcut(s)
		}
	}
	newItem := fyne.NewMenuItem("Compose Message", func() {
		w := a.NewWindow("New Message")
		w.SetContent(views.MakeMessageComposer(w, *identity, *contacts, *sender))
		w.Show()
	})
	settingsItem := fyne.NewMenuItem("Settings", func() {
		w := a.NewWindow("Fyne Settings")
		w.SetContent(settings.NewSettings().LoadAppearanceScreen(w))
		w.Resize(fyne.NewSize(480, 480))
		w.Show()
	})

	cutItem := fyne.NewMenuItem("Cut", func() {
		shortcutFocused(&fyne.ShortcutCut{
			Clipboard: (*w).Clipboard(),
		}, *w)
	})
	copyItem := fyne.NewMenuItem("Copy", func() {
		shortcutFocused(&fyne.ShortcutCopy{
			Clipboard: (*w).Clipboard(),
		}, *w)
	})
	pasteItem := fyne.NewMenuItem("Paste", func() {
		shortcutFocused(&fyne.ShortcutPaste{
			Clipboard: (*w).Clipboard(),
		}, *w)
	})
	// TODO implement find
	findItem := fyne.NewMenuItem("Find", func() {})

	helpMenu := fyne.NewMenu("Help") // TODO implement help menu
	//fyne.NewMenuItem("Documentation", func() {
	//	u, _ := url.Parse("https://developer.fyne.io")
	//	_ = (*a).OpenURL(u)
	//}),
	//fyne.NewMenuItem("Support", func() {
	//	u, _ := url.Parse("https://fyne.io/support/")
	//	_ = (*a).OpenURL(u)
	//}),
	//fyne.NewMenuItemSeparator(),
	//fyne.NewMenuItem("Sponsor", func() {
	//	u, _ := url.Parse("https://github.com/sponsors/fyne-io")
	//	_ = (*a).OpenURL(u)
	//})

	mainMenu := fyne.NewMainMenu(
		// a quit item will be appended to our first menu
		fyne.NewMenu("File", newItem, fyne.NewMenuItemSeparator(), settingsItem),
		fyne.NewMenu("Edit", cutItem, copyItem, pasteItem, fyne.NewMenuItemSeparator(), findItem),
		helpMenu,
	)
	(*w).SetMainMenu(mainMenu)
}

func Run(ipfs *ipmail.Ipfs, sender ipmail.Sender, receiver ipmail.Receiver,
	identity crypto.SelfIdentity, contacts crypto.ContactsIdentityList,
	messages ipmail.MessageList, sent ipmail.MessageList, requests ipmail.MessageList) {

	if messages == nil {
		messages = ipmail.NewMessageList()
	}
	if sent == nil {
		sent = ipmail.NewMessageList()
	}
	if requests == nil {
		requests = ipmail.NewMessageList()
	}

	a := app.NewWithID("io.libipmail")
	topWindow := a.NewWindow("InterPlanetary Mail")
	initMenuBar(&topWindow, &identity, &contacts, &sender)
	topWindow.SetMaster()
	inboxView := views.MakeContent(messages)
	sentView := views.MakeContent(sent)
	var content *container.Split
	setWindowContentTo := func(object fyne.CanvasObject) func() {
		return func() {
			content.Trailing = object
			content.Refresh()
		}
	}
	toolbar := widget.NewToolbar(widget.NewToolbarSpacer())
	content = container.NewHSplit(
		views.MakeNav(map[string]views.SelectFunction{
			"Inbox": setWindowContentTo(inboxView),
			"Sent":  setWindowContentTo(sentView),
		}), inboxView)
	// TODO add container.NewAppTabs()
	topWindow.SetContent(container.NewBorder(toolbar, nil, nil, nil, content))

	identitySet := sync.Mutex{}
	if identity == nil {
		identitySet.Lock()
		println("Prompting for identity")
		onResults := func(results []string, err error) {
			if err != nil {
				println(err.Error())
				return
			}
			name := results[0]
			comment := results[1]
			email := results[2]
			identity, err = crypto.NewSelfIdentity(name, comment, email)
			if err != nil {
				println(err.Error())
				return
			}
			err = identity.SaveToFile(viper.GetString("identity"))
			if err != nil {
				println(err.Error())
				return
			}
			if contacts != nil {
				contacts.Add(identity.EntityList()...)
			}
			identitySet.Unlock()
		}
		prompt(topWindow, onResults, false, "Welcome",
			"Looks like this is your first time here. Welcome! "+
				"You can optionally enter your name, a comment, and your email "+
				"to help identify yourself to people you message. "+
				"Don't worry this information is only stored on your computer.",
			"Name", "Comment", "Email")
	}

	go func() {
		if contacts == nil {
			identitySet.Lock()
			contacts = crypto.NewContactsIdentityList(identity.EntityList())
			identitySet.Unlock()
			err := contacts.SaveToFile(viper.GetString("contacts"))
			if err != nil {
				panic(err.Error())
			}
		}
		identitySet.Lock()
		identitySet.Unlock()

		toolbar.Append(widget.NewToolbarAction(theme.MailComposeIcon(), func() {
			w := a.NewWindow("New Message")
			w.SetContent(views.MakeMessageComposer(w, identity, contacts, sender))
			w.Show()
		}))

		contactRequests := views.MakeContactRequestsManager(contacts, requests)
		toolbar.Append(widget.NewToolbarAction(theme.VisibilityIcon(), func() {
			w := a.NewWindow("Contact Requests")
			w.SetContent(contactRequests)
			w.Show()
		}))

		contactsHashList := newEntityHashList(contacts.ToArray(), ipfs)
		contactsList := views.MakeContactsList(contacts, contactsHashList)
		toolbar.Append(widget.NewToolbarAction(theme.ComputerIcon(), func() {
			w := a.NewWindow("Contacts List")
			w.SetContent(contactsList)
			w.Show()
		}))

		identityHashList := newEntityHashList(identity.EntityList(), ipfs)

		toolbar.Append(widget.NewToolbarAction(theme.MailSendIcon(), func() {
			self_id := identityHashList.Front().Value.(cid.Cid)
			d := dialog.NewEntryDialog("Your ID is \""+self_id.String()+"\"", "Share to:", func(s string) {
				id, err := cid.Parse(s)
				to := crypto.NewIdentityList(identity.DefaultIdentity())
				if err == nil {
					entityBuf, err := ipfs.Cat(path.IpfsPath(id))
					if err == nil {
						entity, err := gpg.ReadEntity(
							packet.NewReader(
								bytes.NewBuffer(entityBuf),
							),
						)
						if err == nil {
							to.Add(entity) // TODO don't suppress errors
						} else {
							println(err.Error())
							return
						}
					} else {
						println(err.Error())
						return
					}
				} else {
					println(err.Error())
					return
				}
				content := bytes.NewBuffer(append([]byte("ipfs:"), self_id.Bytes()...))
				_, err = sender.Send(content, true, identity.DefaultIdentity(), to.ToArray()...)
			}, topWindow)
			d.Show()
		}))

		receiver.OnMessage(func(message iface.PubSubMessage) {
			seq, _ := util.BytesToUint64(message.Seq())
			origin := message.From()
			hash := message.Data()
			topics := strings.Join(message.Topics(), "")
			if strings.Contains(topics, crypto.MessageTopicName) {
				if bytes.HasPrefix(hash, []byte(crypto.MessageCidPrefix)) &&
					bytes.HasSuffix(hash, []byte(crypto.MessageCidPostfix)) {
					hash = bytes.TrimPrefix(hash, []byte(crypto.MessageCidPrefix))
					hash = bytes.TrimSuffix(hash, []byte(crypto.MessageCidPostfix))
					parse, err := cid.Parse(hash)
					if err != nil {
						return
					}
					encryptedMsg, err := ipfs.Cat(path.IpfsPath(parse))
					if err != nil {
						return
					}
					msg := crypto.NewMessage(encryptedMsg, seq, origin, ipfs, identity, contacts,
						func(keys []gpg.Key, symmetric bool) ([]byte, error) {
							result := make([]byte, 0)
							keyStrings := make([]string, 0)
							for _, key := range keys {
								keyStrings = append(keyStrings, key.PublicKey.KeyIdShortString())
							}
							resultMtx := sync.Mutex{}
							resultMtx.Lock()
							onResults := func(results []string, err error) {
								for _, str := range results {
									result = append(result, str...)
								}
								resultMtx.Unlock()
							}
							prompt(topWindow, onResults, true,
								"Unlock your encrypted private keys", "", keyStrings...)
							resultMtx.Lock()
							defer resultMtx.Unlock()
							return result, err
						})
					if msg != nil {
						entities := identity.EntityList()
						for _, entity := range entities { // if msg is from self add to sent list
							if msg.IsFrom(entity) {
								sent.Add(msg)
								err := sent.SaveToFile(viper.GetString("sent"))
								if err != nil {
									println("warning: sent messages could not be saved to file due to:", err.Error())
								}
								return
							}
						}

						entities = contacts.ToArray()
						inContacts := false
						for _, entity := range entities { // if msg is from a non contact add to requests list
							if msg.IsFrom(entity) {
								inContacts = true
								break
							}
						}
						if !inContacts {
							requests.Add(msg)
							err := messages.SaveToFile(viper.GetString("requests"))
							if err != nil {
								println("warning: contact requests could not be saved to file due to:", err.Error())
							}
							contactRequests.Refresh()
							return
						}

						messages.Add(msg)
						err := messages.SaveToFile(viper.GetString("messages"))
						if err != nil {
							println("warning: received messages could not be saved to file due to:", err.Error())
						}
					} else {
						//println("Message isn't meant for you")
					}
				}
			}
		}, true)
	}()

	topWindow.ShowAndRun()

	//print("==> ")
	//for false {
	//	read := ""
	//	sharing := false
	//send:
	//	if strings.HasPrefix(read, "send ") {
	//		to := crypto.NewIdentityList()
	//		trimmed := strings.TrimPrefix(read, "send ")
	//		split := strings.Split(trimmed, " ")
	//		removed := 0
	//		for i, v := range split {
	//			if strings.HasPrefix(v, "to:") {
	//				v = strings.TrimPrefix(v, "to:")
	//				found := contacts.GetByName(v).ToArray()
	//				//for _, val := range found {
	//				//	println(val, "is a recipient candidate")
	//				//}
	//				if len(found) > 1 {
	//					println("More than one contact found:")
	//					array, err := chooseFromArray("Send to who?", func() string {
	//						return ""
	//					}, found, util.EntityToString) // TODO print the IPFS CID hashes
	//					if err != nil {
	//						println("warning:", err)
	//						found = found[:0]
	//					} else {
	//						found = array
	//					}
	//				}
	//				//for _, val := range found {
	//				//	println(val, "was added to the recipient list")
	//				//}
	//				to.Add(found...)
	//				split = append(split[:i-removed], split[i-removed+1:]...)
	//				removed++
	//			} else if sharing && strings.HasPrefix(v, "ipfsto:") { // FIXME EOF error on message received ---- fixed?
	//				v = strings.TrimPrefix(v, "ipfsto:") // FIXME new "unexpected EOF" ---- 2) fixed by x/crypto fork
	//				id, err := cid.Parse(v)
	//				if err == nil {
	//					entityBuf, err := ipfs.Cat(path.IpfsPath(id))
	//					if err == nil {
	//						entity, err := gpg.ReadEntity(
	//							packet.NewReader(
	//								bytes.NewBuffer(entityBuf),
	//							),
	//						)
	//						if err == nil {
	//							to.Add(entity)
	//						} else {
	//							println(err.Error())
	//						}
	//					} else {
	//						println(err.Error())
	//					}
	//				} else {
	//					println(err.Error())
	//				}
	//				split = append(split[:i-removed], split[i-removed+1:]...)
	//				removed++
	//			}
	//		}
	//		toArr := to.ToArray()
	//		if len(toArr) > 0 {
	//			toSend := bytes.NewBufferString(strings.Join(split, " "))
	//			send, err := sender.Send(toSend, true, identity.DefaultIdentity(), append(toArr, identity.DefaultIdentity())...)
	//			if err != nil {
	//				println(err)
	//			} else {
	//				fmt.Printf("Message Sent with CID: %s\n", send)
	//			}
	//		} else {
	//			fmt.Println("Message has no Recipient")
	//		}
	//	} else if strings.HasPrefix(read, "list") {
	//		read = strings.TrimSpace(read[4:])
	//		if strings.HasPrefix(read, "sent") {
	//			println("---- Sent ----")
	//			sent.ForEach(func(message crypto.Message) {
	//				println(message.String())
	//			})
	//		} else {
	//			println("--- Inbox ----")
	//			messages.ForEach(func(message crypto.Message) {
	//				println(message.String())
	//			})
	//		}
	//	} else if strings.HasPrefix(read, "read ") {
	//		trimmed := strings.TrimSpace(read[5:])
	//		reading := messages
	//		split := strings.Split(trimmed, " ")
	//		if strings.EqualFold(split[0], "sent") {
	//			reading = sent
	//			split = split[1:]
	//		}
	//		id, err := strconv.ParseUint(split[0], 10, 64)
	//		if err != nil {
	//			println(err)
	//		} else {
	//			msg := reading.FromId(id)
	//			if msg != nil {
	//				fmt.Printf("%s\n%s\n", msg.String(), msg.Data())
	//			} else {
	//				fmt.Println("Could not find message")
	//			}
	//		}
	//	} else if strings.HasPrefix(read, "contacts ") {
	//		read = strings.TrimPrefix(read, "contacts ")
	//		if strings.HasPrefix(read, "add ") {
	//			read = strings.TrimPrefix(read, "add ")
	//			println("Parsing entity")
	//			go func() {
	//				entity, err := util.ParseEntity(read, ipfs)
	//				println("Finished parsing entity")
	//				if err != nil {
	//					fmt.Printf("\"%s\" is not a valid entity\n", read)
	//				} else {
	//					resolved, err := func() (path.Resolved, error) {
	//						r, w := io.Pipe()
	//						defer r.Close()
	//						go func() {
	//							defer w.Close()
	//							entity.Serialize(w)
	//						}()
	//						return ipfs.AddFromReader(r)
	//					}()
	//					if err != nil {
	//						contactsHashList.PushBack(nil)
	//						println("warning: entity not added to IPFS:", err)
	//					} else {
	//						contactsHashList.PushBack(resolved.Cid())
	//					}
	//					contacts.Add(entity)
	//					err = contacts.SaveToFile(viper.GetString("contacts"))
	//					if err != nil {
	//						fmt.Println("warning: contacts could not be saved to file due to:", err.Error())
	//					}
	//					fmt.Println("Added", util.EntityToString(entity), "to contacts")
	//					print("==> ")
	//				}
	//			}()
	//		} else if strings.HasPrefix(read, "list") || len(strings.TrimSpace(read)) == 0 {
	//			printEntities(read, contacts.ToArray(), contactsHashList)
	//		} else if strings.HasPrefix(read, "requests") {
	//			read := strings.TrimSpace(read[8:])
	//			if strings.HasPrefix(read, "accept") {
	//				read = strings.TrimSpace(read[6:])
	//				split := strings.Split(read, " ")
	//				for _, toAccept := range split {
	//					id, err := strconv.ParseUint(toAccept, 10, 64)
	//					if err != nil {
	//						fmt.Println("warning: could not parse", toAccept, " due to:", err.Error())
	//					} else {
	//						msg := requests.FromId(id)
	//						contacts.Add(msg.From())
	//						requests.Remove(msg)
	//					}
	//				}
	//			} else if strings.HasPrefix(read, "deny") {
	//				read = strings.TrimSpace(read[4:])
	//				split := strings.Split(read, " ")
	//				for _, toDeny := range split {
	//					id, err := strconv.ParseUint(toDeny, 10, 64)
	//					if err != nil {
	//						fmt.Println("warning: could not parse", toDeny, " due to:", err.Error())
	//					} else {
	//						msg := requests.FromId(id)
	//						requests.Remove(msg)
	//					}
	//				}
	//			} else {
	//				println("-- Requests --")
	//				requests.ForEach(func(message crypto.Message) {
	//					println(message.String())
	//				})
	//			}
	//		}
	//	} else if strings.HasPrefix(read, "identity") {
	//		read = strings.TrimSpace(read[8:])
	//		if strings.HasPrefix(read, "share") {
	//			read = strings.TrimSpace(read[5:])
	//			id := identityHashList.Front().Value.(cid.Cid)
	//			read = "send ipfsto:" + read + " ipfs:" + id.String()
	//			sharing = true
	//			goto send
	//		} else {
	//			printEntities(read, identity.EntityList(), identityHashList)
	//		}
	//	} else if strings.HasPrefix(read, "exit") ||
	//		strings.HasPrefix(read, "quit") {
	//		return
	//	} else if strings.HasPrefix(read, "help") ||
	//		strings.HasPrefix(read, "?") {
	//		println("-------- Commands --------\n")
	//		println("help - Prints out this message")
	//		println("? - Prints out this message")
	//		println("contacts [list] - Prints a list of your contacts")
	//		println("contacts add <content ID> - Tries to add a contact by their content ID")
	//		println("contacts requests - Prints a list of your contact requests")
	//		println("contacts requests [accept|deny] <request ID> - Accepts or denies a contact request")
	//		println("exit - Quits the mail client")
	//		println("identity - Prints an IPFS content ID for your default identity")
	//		println("identity qrcode - Prints a QR code of your default identity")
	//		println("identity share <content ID> - Shares your identity with anyone by their content ID")
	//		println("list - Prints a summary of your received messages")
	//		println("list sent - Prints a summary of all your sent messages")
	//		println("quit - Quits the mail client")
	//		println("read <message ID> - Prints out a received message with a given message ID")
	//		println("read sent <message ID> - Prints out a sent message with a given message ID")
	//		println("send [to:<contact name>] [ipfsto:<contact content ID>] <message>")
	//		println("        Sends a message to recipients listed by to and ipfsto arguments using")
	//		println("        the contact's name and contact's content ID, respectively. There can")
	//		println("        be as many to and ipfsto arguments as you like and they can even be")
	//		println("        in the message and collected, so be careful not to start a word with")
	//		println("        \"to:\" or \"ipfsto:\".")
	//	}
	//	print("==> ")
	//}

}

func newEntityHashList(entities gpg.EntityList, ipfs *ipmail.Ipfs) *list.List {
	identityHashList := list.New()
	for _, entity := range entities {
		func() {
			r, w := io.Pipe()
			defer r.Close()
			go func() {
				defer w.Close()
				entity.Serialize(w)
			}()
			resolved, _ := ipfs.AddFromReader(r)
			if resolved != nil {
				identityHashList.PushBack(resolved.Cid())
			} else {
				identityHashList.PushBack(nil)
			}
		}()
	}
	return identityHashList
}

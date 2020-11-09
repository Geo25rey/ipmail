package cli

import (
	"bufio"
	"bytes"
	"container/list"
	"errors"
	"fmt"
	gpg "github.com/Geo25rey/crypto/openpgp"
	"github.com/Geo25rey/crypto/openpgp/packet"
	"github.com/Jguer/yay/v10/pkg/intrange"
	"github.com/ipfs/go-cid"
	iface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/skip2/go-qrcode"
	"github.com/spf13/viper"
	"io"
	"ipmail/ipmail"
	"ipmail/ipmail/crypto"
	"ipmail/ipmail/util"
	"os"
	"strconv"
	"strings"
)

func hasInvalidCharacters(s string) (bool, int32) {
	for _, c := range s {
		switch c {
		case '(', ')', '<', '>', 0:
			return true, c
		}
	}
	return false, 0
}

func prompt(scanner *bufio.Scanner, field string) string {
	for true {
		fmt.Printf("%s?\n> ", field)
		scanner.Scan()
		val := scanner.Text()
		checkFailed, failedChar := hasInvalidCharacters(val)
		if checkFailed {
			fmt.Printf("%s can not contain '%c', try again", field, failedChar)
			continue
		}
		return val
	}
	return ""
}

func Run(ipfs *ipmail.Ipfs, sender ipmail.Sender, receiver ipmail.Receiver,
	identity crypto.SelfIdentity, contacts crypto.ContactsIdentityList,
	messages ipmail.MessageList, sent ipmail.MessageList, requests ipmail.MessageList) {

	scanner := bufio.NewScanner(os.Stdin)
	if messages == nil {
		messages = ipmail.NewMessageList()
	}
	if sent == nil {
		sent = ipmail.NewMessageList()
	}
	if requests == nil {
		requests = ipmail.NewMessageList()
	}

	if identity == nil {
		println("Looks like this is your first time here. Welcome!")
		println("You can optionally enter your name, a comment, and your email")
		println("to help identify yourself to people you message.")
		println("Don't worry this information is only stored on your computer.")
		name := prompt(scanner, "Name")
		comment := prompt(scanner, "Comment")
		email := prompt(scanner, "Email")
		var err error
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
	}

	if contacts == nil {
		contacts = crypto.NewContactsIdentityList(identity.EntityList())
		err := contacts.SaveToFile(viper.GetString("contacts"))
		if err != nil {
			panic(err)
		}
	}

	identityHashList := newEntityHashList(identity.EntityList(), ipfs)
	contactsHashList := newEntityHashList(contacts.ToArray(), ipfs)

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
						for _, key := range keys {
							println("==> End your passphrase for", key.PublicKey.KeyIdShortString())
							print("==> ")
							if !scanner.Scan() {
								return nil, errors.New("EOF")
							}
							result = append(result, scanner.Bytes()...)
						}
						return result, nil
					})
				if msg != nil {
					entities := identity.EntityList()
					for _, entity := range entities { // if msg is from self add to sent list
						if msg.IsFrom(entity) {
							fmt.Println("Message added to sent list")
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
						fmt.Println("Contact Request added")
						requests.Add(msg)
						err := messages.SaveToFile(viper.GetString("requests"))
						if err != nil {
							println("warning: contact requests could not be saved to file due to:", err.Error())
						}
					}

					fmt.Println("Message added to inbox")
					messages.Add(msg)
					err := messages.SaveToFile(viper.GetString("messages"))
					if err != nil {
						println("warning: received messages could not be saved to file due to:", err.Error())
					}
				} else {
					//println("Message isn't meant for you")
				}
				print("==> ")
			}
		}
	}, true)

	print("==> ")
	for scanner.Scan() {
		read := scanner.Text()
		if strings.HasPrefix(read, "send ") {
			to := crypto.NewIdentityList()
			trimmed := strings.TrimPrefix(read, "send ")
			split := strings.Split(trimmed, " ")
			removed := 0
			for i, v := range split {
				if strings.HasPrefix(v, "to:") {
					v = strings.TrimPrefix(v, "to:")
					found := contacts.GetByName(v).ToArray()
					//for _, val := range found {
					//	println(val, "is a recipient candidate")
					//}
					if len(found) > 1 {
						println("More than one contact found:")
						array, err := chooseFromArray("Send to who?", func() string {
							scanner.Scan()
							return scanner.Text()
						}, found, util.EntityToString) // TODO print the IPFS CID hashes
						if err != nil {
							println("warning:", err)
							found = found[:0]
						} else {
							found = array
						}
					}
					//for _, val := range found {
					//	println(val, "was added to the recipient list")
					//}
					to.Add(found...)
					split = append(split[:i-removed], split[i-removed+1:]...)
					removed++
				}
				if strings.HasPrefix(v, "ipfsto:") { // FIXME EOF error on message received ---- fixed?
					v = strings.TrimPrefix(v, "ipfsto:") // FIXME new "unexpected EOF" ---- 2) fixed by x/crypto fork
					id, err := cid.Parse(v)
					if err == nil {
						entityBuf, err := ipfs.Cat(path.IpfsPath(id))
						if err == nil {
							entity, err := gpg.ReadEntity(
								packet.NewReader(
									bytes.NewBuffer(entityBuf),
								),
							)
							if err == nil {
								to.Add(entity)
							} else {
								println(err.Error())
							}
						} else {
							println(err.Error())
						}
					} else {
						println(err.Error())
					}
					split = append(split[:i-removed], split[i-removed+1:]...)
					removed++
				}
			}
			toArr := to.ToArray()
			if len(toArr) > 0 {
				toSend := bytes.NewBufferString(strings.Join(split, " "))
				send, err := sender.Send(toSend, true, identity.DefaultIdentity(), append(toArr, identity.DefaultIdentity())...)
				if err != nil {
					println(err)
				} else {
					fmt.Printf("Message Sent with CID: %s\n", send)
				}
			} else {
				fmt.Println("Message has no Recipient")
			}
		} else if strings.HasPrefix(read, "list") {
			read = strings.TrimSpace(read[4:])
			if strings.HasPrefix(read, "sent") {
				println("---- Sent ----")
				sent.ForEach(func(message crypto.Message) {
					println(message.String())
				})
			} else {
				println("--- Inbox ----")
				messages.ForEach(func(message crypto.Message) {
					println(message.String())
				})
			}
		} else if strings.HasPrefix(read, "read ") {
			trimmed := strings.TrimSpace(read[5:])
			reading := messages
			split := strings.Split(trimmed, " ")
			if strings.EqualFold(split[0], "sent") {
				reading = sent
				split = split[1:]
			}
			id, err := strconv.ParseUint(split[0], 10, 64)
			if err != nil {
				println(err)
			} else {
				msg := reading.FromId(id)
				if msg != nil {
					fmt.Printf("%s\n%s\n", msg.String(), msg.Data())
				} else {
					fmt.Println("Could not find message")
				}
			}
		} else if strings.HasPrefix(read, "contacts ") {
			read = strings.TrimPrefix(read, "contacts ")
			if strings.HasPrefix(read, "add ") {
				read = strings.TrimPrefix(read, "add ")
				println("Parsing entity")
				go func() {
					entity, err := util.ParseEntity(read, ipfs)
					println("Finished parsing entity")
					if err != nil {
						fmt.Printf("\"%s\" is not a valid entity\n", read)
					} else {
						resolved, err := func() (path.Resolved, error) {
							r, w := io.Pipe()
							defer r.Close()
							go func() {
								defer w.Close()
								entity.Serialize(w)
							}()
							return ipfs.AddFromReader(r)
						}()
						if err != nil {
							contactsHashList.PushBack(nil)
							println("warning: entity not added to IPFS:", err)
						} else {
							contactsHashList.PushBack(resolved.Cid())
						}
						contacts.Add(entity)
						err = contacts.SaveToFile(viper.GetString("contacts"))
						if err != nil {
							fmt.Println("warning: contacts could not be saved to file due to:", err.Error())
						}
						fmt.Println("Added", util.EntityToString(entity), "to contacts")
						print("==> ")
					}
				}()
			} else if strings.HasPrefix(read, "list") || len(strings.TrimSpace(read)) == 0 {
				printEntities(read, contacts.ToArray(), contactsHashList)
			} else if strings.HasPrefix(read, "requests") {
				read := strings.TrimSpace(read[8:])
				if strings.HasPrefix(read, "accept") {
					read = strings.TrimSpace(read[6:])
					split := strings.Split(read, " ")
					for _, toAccept := range split {
						id, err := strconv.ParseUint(toAccept, 10, 64)
						if err != nil {
							fmt.Println("warning: could not parse", toAccept, " due to:", err.Error())
						} else {
							msg := requests.FromId(id)
							contacts.Add(msg.From())
							requests.Remove(msg)
						}
					}
				} else if strings.HasPrefix(read, "deny") {
					read = strings.TrimSpace(read[4:])
					split := strings.Split(read, " ")
					for _, toDeny := range split {
						id, err := strconv.ParseUint(toDeny, 10, 64)
						if err != nil {
							fmt.Println("warning: could not parse", toDeny, " due to:", err.Error())
						} else {
							msg := requests.FromId(id)
							requests.Remove(msg)
						}
					}
				} else {
					println("-- Requests --")
					requests.ForEach(func(message crypto.Message) {
						println(message.String())
					})
				}
			}
		} else if strings.HasPrefix(read, "identity") {
			read = strings.TrimSpace(read[8:])
			if strings.HasPrefix(read, "share") {
				read = strings.TrimSpace(read[5:])
				id := identityHashList.Front().Value.(cid.Cid)
				read = "send ipfsto:" + read + " ipfs:" + id.String()
				sharing = true
				goto send
			} else {
				printEntities(read, identity.EntityList(), identityHashList)
			}
		} else if strings.HasPrefix(read, "exit") ||
			strings.HasPrefix(read, "quit") {
			return
		} else if strings.HasPrefix(read, "help") ||
			strings.HasPrefix(read, "?") {
			println("-------- Commands --------\n")
			println("help - Prints out this message")
			println("? - Prints out this message")
			println("contacts [list] - Prints a list of your contacts")
			println("contacts add <content ID> - Tries to add a contact by their content ID")
			println("contacts requests - Prints a list of your contact requests")
			println("contacts requests [accept|deny] <request ID> - Accepts or denies a contact request")
			println("exit - Quits the mail client")
			println("identity - Prints an IPFS content ID for your default identity")
			println("identity qrcode - Prints a QR code of your default identity")
			println("identity share <content ID> - Shares your identity with anyone by their content ID")
			println("list - Prints a summary of your received messages")
			println("list sent - Prints a summary of all your sent messages")
			println("quit - Quits the mail client")
			println("read <message ID> - Prints out a received message with a given message ID")
			println("read sent <message ID> - Prints out a sent message with a given message ID")
			println("send [to:<contact name>] [ipfsto:<contact content ID>] <message>")
			println("        Sends a message to recipients listed by to and ipfsto arguments using")
			println("        the contact's name and contact's content ID, respectively. There can")
			println("        be as many to and ipfsto arguments as you like and they can even be")
			println("        in the message and collected, so be careful not to start a word with")
			println("        \"to:\" or \"ipfsto:\".")
		}
		print("==> ")
	}
}

func chooseFromArray(prompt string, input func() string, array []*gpg.Entity, toString func(entity *gpg.Entity) string) ([]*gpg.Entity, error) {
	for i, entity := range array {
		println("", "", i+1, toString(entity))
	}
	println("==>", prompt)
	println("==> [N]one [A]ll or (1 2 3, 1-3, ^4)")
	print("==> ")

	eInclude, eExclude, eOtherInclude, eOtherExclude := intrange.ParseNumberMenu(input())
	eIsInclude := len(eExclude) == 0 && len(eOtherExclude) == 0

	if eOtherInclude.Get("abort") || eOtherInclude.Get("ab") {
		return nil, fmt.Errorf("aborting due to user")
	}

	toEdit := make([]*gpg.Entity, 0)
	if !eOtherInclude.Get("n") && !eOtherInclude.Get("none") {
		for i, base := range array {
			if !eIsInclude && eExclude.Get(len(array)-i-1) {
				continue
			}

			if eOtherInclude.Get("a") || eOtherInclude.Get("all") {
				toEdit = append(toEdit, base)
				continue
			}

			if eIsInclude && (eInclude.Get(len(array) - i - 1)) {
				toEdit = append(toEdit, base)
			}

			if !eIsInclude && (!eExclude.Get(len(array) - i - 1)) {
				toEdit = append(toEdit, base)
			}
		}
	}

	return toEdit, nil
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

func printEntities(read string, entities gpg.EntityList, hashList *list.List) {
	printQR := false
	if strings.EqualFold(read, "qrcode") {
		printQR = true
	}
	for i, hash := 0, hashList.Front(); i < len(entities) && hash != nil; i, hash = i+1, hash.Next() {
		if hash.Value == nil {
			continue
		}
		id := hash.Value.(cid.Cid)
		entity := entities[i]
		entityStr := util.EntityToString(entity)
		qrStr := ""
		if printQR {
			qr, err := qrcode.New(id.String(), qrcode.Low)
			if err != nil {
				continue
			}
			qrStr = qr.ToSmallString(false)
		}
		fmt.Printf("%s%s -> %s\n", qrStr, id, entityStr)
	}
}

package crypto

import (
	"bytes"
	"container/list"
	gpg "github.com/Geo25rey/crypto/openpgp"
	"github.com/Geo25rey/crypto/openpgp/packet"
	"github.com/libp2p/go-libp2p-core/peer"
	"io"
	"ipmail/libipmail/util"
	"reflect"
	"testing"
)

var (
	entity1, _ = gpg.NewEntity("name 1", "comment 1", "email 1", util.DefaultEncryptionConfig())
	entity2, _ = gpg.NewEntity("name 2", "comment 2", "email 2", util.DefaultEncryptionConfig())
)

func TestNewContactsIdentityList(t *testing.T) {
	type args struct {
		entities gpg.EntityList
	}
	tests := []struct {
		name string
		args args
		want ContactsIdentityList
	}{
		{
			"Empty List",
			args{
				[]*gpg.Entity{},
			},
			&contactsIdentityList{
				NewIdentityList(),
			},
		},
		{
			"List with nil",
			args{
				[]*gpg.Entity{
					nil,
				},
			},
			&contactsIdentityList{
				NewIdentityList(),
			},
		},
		{
			"Normal List",
			args{
				[]*gpg.Entity{
					entity1,
					entity2,
				},
			},
			&contactsIdentityList{
				NewIdentityList(entity1, entity2),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewContactsIdentityList(tt.args.entities); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewContactsIdentityList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewContactsIdentityListFromFile(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name    string
		args    args
		want    ContactsIdentityList
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewContactsIdentityListFromFile(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewContactsIdentityListFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewContactsIdentityListFromFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func NewListWith(these ...interface{}) (result *list.List) {
	result = list.New()
	for _, this := range these {
		result.PushBack(this)
	}
	return
}

func TestNewIdentityList1(t *testing.T) {
	type args struct {
		entities []*gpg.Entity
	}
	tests := []struct {
		name string
		args args
		want IdentityList
	}{
		{
			"Empty List",
			args{
				[]*gpg.Entity{},
			},
			&identityList{
				list.New(),
			},
		},
		{
			"List with nil",
			args{
				[]*gpg.Entity{nil},
			},
			&identityList{
				list.New(),
			},
		},
		{
			"Normal List",
			args{
				[]*gpg.Entity{
					entity1,
					entity2,
				},
			},
			&identityList{
				NewListWith(entity1, entity2),
			},
		},
		{
			"List with repeats",
			args{
				[]*gpg.Entity{
					entity1,
					entity1,
				},
			},
			&identityList{
				NewListWith(entity1),
			},
		},
		{
			"Normal List with repeats",
			args{
				[]*gpg.Entity{
					entity1,
					entity1,
					entity2,
				},
			},
			&identityList{
				NewListWith(entity1, entity2),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewIdentityList(tt.args.entities...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewIdentityList() = %v, want %v", got.ToArray(), tt.want.ToArray())
			}
		})
	}
}

func TestNewMessage(t *testing.T) {
	type args struct {
		encryptedData []byte
		id            uint64
		origin        peer.ID
		ipfs          util.Cat
		identity      SelfIdentity
		contacts      ContactsIdentityList
		prompt        gpg.PromptFunction
	}
	tests := []struct {
		name string
		args args
		want Message
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMessage(tt.args.encryptedData, tt.args.id, tt.args.origin, tt.args.ipfs, tt.args.identity, tt.args.contacts, tt.args.prompt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSelfIdentity(t *testing.T) {
	type args struct {
		name    string
		comment string
		email   string
	}
	tests := []struct {
		name    string
		args    args
		want    SelfIdentity
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSelfIdentity(tt.args.name, tt.args.comment, tt.args.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSelfIdentity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSelfIdentity() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSelfIdentityFromFile(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want SelfIdentity
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSelfIdentityFromFile(tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSelfIdentityFromFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadMessage(t *testing.T) {
	type args struct {
		r        io.Reader
		ipfs     util.Cat
		identity SelfIdentity
		contacts ContactsIdentityList
	}
	tests := []struct {
		name    string
		args    args
		want    Message
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadMessage(tt.args.r, tt.args.ipfs, tt.args.identity, tt.args.contacts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadMessage() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_contactsIdentityList_SaveToFile(t *testing.T) {
	type fields struct {
		IdentityList IdentityList
	}
	type args struct {
		file string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &contactsIdentityList{
				IdentityList: tt.fields.IdentityList,
			}
			if err := c.SaveToFile(tt.args.file); (err != nil) != tt.wantErr {
				t.Errorf("SaveToFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_identityList_Add(t *testing.T) {
	type fields struct {
		list *list.List
	}
	type args struct {
		entities []*gpg.Entity
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *identityList
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &identityList{
				list: tt.fields.list,
			}
			if i.Add(tt.args.entities...); !reflect.DeepEqual(i, tt.want) {
				t.Errorf("Add() = %v, want %v", i, tt.want)
			}
		})
	}
}

func Test_identityList_AddFromKeyRing(t *testing.T) {
	type fields struct {
		list *list.List
	}
	type args struct {
		ring gpg.KeyRing
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *identityList
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &identityList{
				list: tt.fields.list,
			}
			if i.AddFromKeyRing(tt.args.ring); !reflect.DeepEqual(i, tt.want) {
				t.Errorf("AddFromKeyRing() = %v, want %v", i, tt.want)
			}
		})
	}
}

func Test_identityList_ForEach(t *testing.T) {
	type fields struct {
		list *list.List
	}
	type args struct {
		do func(entity *gpg.Entity)
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		recieved []bool
		want     []bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &identityList{
				list: tt.fields.list,
			}
			if i.ForEach(tt.args.do); !reflect.DeepEqual(tt.recieved, tt.want) {
				t.Errorf("ForEach() = %v, want %v", tt.recieved, tt.want)
			}
		})
	}
}

func Test_identityList_GetAny(t *testing.T) {
	type fields struct {
		list *list.List
	}
	tests := []struct {
		name   string
		fields fields
		want   *gpg.Entity
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &identityList{
				list: tt.fields.list,
			}
			if got := i.GetAny(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAny() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_identityList_GetByEmail(t *testing.T) {
	type fields struct {
		list *list.List
	}
	type args struct {
		email string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   IdentityList
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &identityList{
				list: tt.fields.list,
			}
			if got := i.GetByEmail(tt.args.email); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetByEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_identityList_GetByName(t *testing.T) {
	type fields struct {
		list *list.List
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   IdentityList
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &identityList{
				list: tt.fields.list,
			}
			if got := i.GetByName(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetByName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_identityList_GetByPublicKey(t *testing.T) {
	type fields struct {
		list *list.List
	}
	type args struct {
		key packet.PublicKey
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    IdentityList
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &identityList{
				list: tt.fields.list,
			}
			got, err := i.GetByPublicKey(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByPublicKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetByPublicKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_identityList_ToArray(t *testing.T) {
	type fields struct {
		list *list.List
	}
	tests := []struct {
		name   string
		fields fields
		want   []*gpg.Entity
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &identityList{
				list: tt.fields.list,
			}
			if got := i.ToArray(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_message_Data(t *testing.T) {
	type fields struct {
		encryptedData []byte
		decryptedData []byte
		from          *packet.UserId
		fromEntity    *gpg.Entity
		id            uint64
		origin        peer.ID
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &message{
				encryptedData: tt.fields.encryptedData,
				decryptedData: tt.fields.decryptedData,
				from:          tt.fields.from,
				fromEntity:    tt.fields.fromEntity,
				id:            tt.fields.id,
				origin:        tt.fields.origin,
			}
			if got := m.Data(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Data() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_message_From(t *testing.T) {
	type fields struct {
		encryptedData []byte
		decryptedData []byte
		from          *packet.UserId
		fromEntity    *gpg.Entity
		id            uint64
		origin        peer.ID
	}
	tests := []struct {
		name   string
		fields fields
		want   *gpg.Entity
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &message{
				encryptedData: tt.fields.encryptedData,
				decryptedData: tt.fields.decryptedData,
				from:          tt.fields.from,
				fromEntity:    tt.fields.fromEntity,
				id:            tt.fields.id,
				origin:        tt.fields.origin,
			}
			if got := m.From(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("From() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_message_FromEmail(t *testing.T) {
	type fields struct {
		encryptedData []byte
		decryptedData []byte
		from          *packet.UserId
		fromEntity    *gpg.Entity
		id            uint64
		origin        peer.ID
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &message{
				encryptedData: tt.fields.encryptedData,
				decryptedData: tt.fields.decryptedData,
				from:          tt.fields.from,
				fromEntity:    tt.fields.fromEntity,
				id:            tt.fields.id,
				origin:        tt.fields.origin,
			}
			if got := m.FromEmail(); got != tt.want {
				t.Errorf("FromEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_message_FromName(t *testing.T) {
	type fields struct {
		encryptedData []byte
		decryptedData []byte
		from          *packet.UserId
		fromEntity    *gpg.Entity
		id            uint64
		origin        peer.ID
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &message{
				encryptedData: tt.fields.encryptedData,
				decryptedData: tt.fields.decryptedData,
				from:          tt.fields.from,
				fromEntity:    tt.fields.fromEntity,
				id:            tt.fields.id,
				origin:        tt.fields.origin,
			}
			if got := m.FromName(); got != tt.want {
				t.Errorf("FromName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_message_Id(t *testing.T) {
	type fields struct {
		encryptedData []byte
		decryptedData []byte
		from          *packet.UserId
		fromEntity    *gpg.Entity
		id            uint64
		origin        peer.ID
	}
	tests := []struct {
		name   string
		fields fields
		want   uint64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &message{
				encryptedData: tt.fields.encryptedData,
				decryptedData: tt.fields.decryptedData,
				from:          tt.fields.from,
				fromEntity:    tt.fields.fromEntity,
				id:            tt.fields.id,
				origin:        tt.fields.origin,
			}
			if got := m.Id(); got != tt.want {
				t.Errorf("Id() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_message_IsFrom(t *testing.T) {
	type fields struct {
		encryptedData []byte
		decryptedData []byte
		from          *packet.UserId
		fromEntity    *gpg.Entity
		id            uint64
		origin        peer.ID
	}
	type args struct {
		entity *gpg.Entity
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &message{
				encryptedData: tt.fields.encryptedData,
				decryptedData: tt.fields.decryptedData,
				from:          tt.fields.from,
				fromEntity:    tt.fields.fromEntity,
				id:            tt.fields.id,
				origin:        tt.fields.origin,
			}
			if got := m.IsFrom(tt.args.entity); got != tt.want {
				t.Errorf("IsFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_message_Serialize(t *testing.T) {
	type fields struct {
		encryptedData []byte
		decryptedData []byte
		from          *packet.UserId
		fromEntity    *gpg.Entity
		id            uint64
		origin        peer.ID
	}
	tests := []struct {
		name    string
		fields  fields
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &message{
				encryptedData: tt.fields.encryptedData,
				decryptedData: tt.fields.decryptedData,
				from:          tt.fields.from,
				fromEntity:    tt.fields.fromEntity,
				id:            tt.fields.id,
				origin:        tt.fields.origin,
			}
			w := &bytes.Buffer{}
			err := m.Serialize(w)
			if (err != nil) != tt.wantErr {
				t.Errorf("Serialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Serialize() gotW = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func Test_message_String(t *testing.T) {
	type fields struct {
		encryptedData []byte
		decryptedData []byte
		from          *packet.UserId
		fromEntity    *gpg.Entity
		id            uint64
		origin        peer.ID
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &message{
				encryptedData: tt.fields.encryptedData,
				decryptedData: tt.fields.decryptedData,
				from:          tt.fields.from,
				fromEntity:    tt.fields.fromEntity,
				id:            tt.fields.id,
				origin:        tt.fields.origin,
			}
			if got := m.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_message_decrypt(t *testing.T) {
	type fields struct {
		encryptedData []byte
		decryptedData []byte
		from          *packet.UserId
		fromEntity    *gpg.Entity
		id            uint64
		origin        peer.ID
	}
	type args struct {
		ipfs     util.Cat
		identity SelfIdentity
		contacts ContactsIdentityList
		prompt   gpg.PromptFunction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &message{
				encryptedData: tt.fields.encryptedData,
				decryptedData: tt.fields.decryptedData,
				from:          tt.fields.from,
				fromEntity:    tt.fields.fromEntity,
				id:            tt.fields.id,
				origin:        tt.fields.origin,
			}
			if err := m.decrypt(tt.args.ipfs, tt.args.identity, tt.args.contacts, tt.args.prompt); (err != nil) != tt.wantErr {
				t.Errorf("decrypt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_selfIdentity_DefaultIdentity(t *testing.T) {
	type fields struct {
		identities      IdentityList
		defaultIdentity *gpg.Entity
	}
	tests := []struct {
		name   string
		fields fields
		want   *gpg.Entity
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &selfIdentity{
				identities:      tt.fields.identities,
				defaultIdentity: tt.fields.defaultIdentity,
			}
			if got := s.DefaultIdentity(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultIdentity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_selfIdentity_EntityList(t *testing.T) {
	type fields struct {
		identities      IdentityList
		defaultIdentity *gpg.Entity
	}
	tests := []struct {
		name   string
		fields fields
		want   gpg.EntityList
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &selfIdentity{
				identities:      tt.fields.identities,
				defaultIdentity: tt.fields.defaultIdentity,
			}
			if got := s.EntityList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EntityList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_selfIdentity_SaveToFile(t *testing.T) {
	type fields struct {
		identities      IdentityList
		defaultIdentity *gpg.Entity
	}
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &selfIdentity{
				identities:      tt.fields.identities,
				defaultIdentity: tt.fields.defaultIdentity,
			}
			if err := s.SaveToFile(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("SaveToFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

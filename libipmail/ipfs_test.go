package ipmail

import (
	"bytes"
	"context"
	"github.com/ipfs/go-cid"
	files "github.com/ipfs/go-ipfs-files"
	iface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestIpfs_Add(t *testing.T) {
	type fields struct {
		api iface.CoreAPI
		ctx context.Context
	}
	type args struct {
		node files.Node
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    path.Resolved
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &Ipfs{
				api: tt.fields.api,
				ctx: tt.fields.ctx,
			}
			got, err := this.Add(tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Add() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIpfs_AddFromBytes(t *testing.T) {
	type fields struct {
		api iface.CoreAPI
		ctx context.Context
	}
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    path.Resolved
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &Ipfs{
				api: tt.fields.api,
				ctx: tt.fields.ctx,
			}
			got, err := this.AddFromBytes(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddFromBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddFromBytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIpfs_AddFromPath(t *testing.T) {
	type fields struct {
		api iface.CoreAPI
		ctx context.Context
	}
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    path.Resolved
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &Ipfs{
				api: tt.fields.api,
				ctx: tt.fields.ctx,
			}
			got, err := this.AddFromPath(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddFromPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddFromPath() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIpfs_AddFromReader(t *testing.T) {
	type fields struct {
		api iface.CoreAPI
		ctx context.Context
	}
	type args struct {
		reader io.Reader
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    path.Resolved
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &Ipfs{
				api: tt.fields.api,
				ctx: tt.fields.ctx,
			}
			got, err := this.AddFromReader(tt.args.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddFromReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddFromReader() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIpfs_Cat(t *testing.T) {
	stdoutBuf, stdoutClose := captureOutput(&os.Stdout, false)
	stderrBuf, stderrClose := captureOutput(&os.Stderr, false)
	this, err := NewIpfs(false)
	if err != nil {
		t.Errorf("Failed to create IPFS node")
		t.FailNow()
	}
	testData := []byte("Test Data")
	buf := bytes.NewBuffer(testData)
	cid1, err := this.AddFromReader(buf)
	if err != nil {
		t.Errorf("Failed to add test data 1 to IPFS node")
		t.FailNow()
	}
	type args struct {
		cidFile path.Resolved
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			// TODO Add better checking for IPFS log output
			"Cat File DNE",
			args{
				path.IpfsPath(cid.Undef),
			},
			nil,
			true,
		},
		{
			"Cat File exists",
			args{
				cid1,
			},
			testData,
			false,
		},
	}
	// Clear any prior output
	_ = readAll(stdoutBuf)
	_ = readAll(stderrBuf)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := this.Cat(tt.args.cidFile)
			stdoutOutput := readAll(stdoutBuf)
			stderrOutput := readAll(stderrBuf)
			if strings.Compare(string(stdoutOutput), "") != 0 {
				t.Errorf("os.Stdout got = '%v', want '%v'", string(stdoutOutput), "")
				return
			}
			if strings.Compare(string(stderrOutput), "") != 0 {
				t.Errorf("os.Stderr got = '%v', want '%v'", string(stderrOutput), "")
				return
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("Cat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cat() got = %v, want %v", got, tt.want)
			}
		})
	}
	stdoutClose()
	stderrClose()
}

func readAll(buf *bytes.Buffer) []byte {
	return buf.Next(buf.Len())
}

func captureOutput(output **os.File, writeToOrigin ...bool) (*bytes.Buffer, func()) {
	buf := bytes.NewBuffer(make([]byte, 0))
	r, w, _ := os.Pipe()
	orig := *output
	*output = w
	lenArg := len(writeToOrigin)
	go func() {
		arr := []byte{0}
		for read, err := r.Read(arr); read != 0 && err == nil; read, err = r.Read(arr) {
			if lenArg > 0 && writeToOrigin[0] {
				_, err = orig.Write(arr)
				if err != nil {
					return
				}
			}
			_, err = buf.Write(arr)
			if err != nil {
				return
			}
		}
	}()
	return buf, func() {
		_ = orig.Sync()
		*output = orig
	}

}

func TestIpfs_Context(t *testing.T) {
	type fields struct {
		api iface.CoreAPI
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		want   context.Context
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &Ipfs{
				api: tt.fields.api,
				ctx: tt.fields.ctx,
			}
			if got := this.Context(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Context() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIpfs_Ls(t *testing.T) {
	type fields struct {
		api iface.CoreAPI
		ctx context.Context
	}
	type args struct {
		cidFile path.Resolved
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []files.Node
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &Ipfs{
				api: tt.fields.api,
				ctx: tt.fields.ctx,
			}
			got, err := this.Ls(tt.args.cidFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ls() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ls() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIpfs_Publish(t *testing.T) {
	type fields struct {
		api iface.CoreAPI
		ctx context.Context
	}
	type args struct {
		topic  string
		toSend []byte
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
			this := &Ipfs{
				api: tt.fields.api,
				ctx: tt.fields.ctx,
			}
			if err := this.Publish(tt.args.topic, tt.args.toSend); (err != nil) != tt.wantErr {
				t.Errorf("Publish() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIpfs_Subscribe(t *testing.T) {
	type fields struct {
		api iface.CoreAPI
		ctx context.Context
	}
	type args struct {
		topic   string
		options []options.PubSubSubscribeOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    iface.PubSubSubscription
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &Ipfs{
				api: tt.fields.api,
				ctx: tt.fields.ctx,
			}
			got, err := this.Subscribe(tt.args.topic, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Subscribe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Subscribe() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewIpfs(t *testing.T) {
	type args struct {
		useLocalNode bool
	}
	tests := []struct {
		name    string
		args    args
		want    *Ipfs
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewIpfs(tt.args.useLocalNode)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewIpfs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewIpfs() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewIpfsWithRepo(t *testing.T) {
	type args struct {
		useLocalNode bool
		path         *string
	}
	tests := []struct {
		name    string
		args    args
		want    *Ipfs
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewIpfsWithRepo(tt.args.useLocalNode, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewIpfsWithRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewIpfsWithRepo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

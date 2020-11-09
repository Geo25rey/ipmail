package ipmail

import (
	"context"
	"fmt"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	config "github.com/ipfs/go-ipfs-config"
	files "github.com/ipfs/go-ipfs-files"
	libp2p "github.com/ipfs/go-ipfs/core/node/libp2p"
	icore "github.com/ipfs/interface-go-ipfs-core"
	icorepath "github.com/ipfs/interface-go-ipfs-core/path"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreapi"
	"github.com/ipfs/go-ipfs/plugin/loader" // This package is needed so that all the preloaded plugins are loaded automatically
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	"github.com/libp2p/go-libp2p-core/peer"
)

/// ------ Setting up the IPFS Repo

func setupPlugins(externalPluginsPath string) error {
	// Load any external plugins if available on externalPluginsPath
	plugins, err := loader.NewPluginLoader(filepath.Join(externalPluginsPath, "plugins"))
	if err != nil {
		return fmt.Errorf("error loading plugins: %s", err)
	}

	// Load preloaded and external plugins
	if err := plugins.Initialize(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	if err := plugins.Inject(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	return nil
}

func createTempRepo(repoPath *string) (*string, error) {
	if repoPath == nil {
		path, err := ioutil.TempDir("", "ipfs-shell")
		if err != nil {
			return nil, fmt.Errorf("failed to get temp dir: %s", err)
		}
		repoPath = &path
	}

	// Create a config with default options and a 2048 bit key
	cfg, err := config.Init(os.Stdout, 4096)
	if err != nil {
		return nil, err
	}

	// Create the repo with the config
	err = fsrepo.Init(*repoPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init ephemeral node: %s", err)
	}

	return repoPath, nil
}

/// ------ Spawning the node

// Creates an IPFS node and returns its coreAPI
func createNode(ctx context.Context, repoPath string) (icore.CoreAPI, error) {
	// Open the repo
	repo, err := fsrepo.Open(repoPath)
	if err != nil {
		return nil, err
	}

	cfg, err := repo.Config()

	// Sets swarm ports to random - helps with port conflicts
	maxTries := 3
	err = config.Profiles["randomports"].Transform(cfg)
	for i := 0; i < maxTries; i++ {
		if err == nil {
			break
		}
		err = config.Profiles["randomports"].Transform(cfg)
	}
	if err != nil {
		return nil, err
	}

	err = repo.SetConfig(cfg)
	if err != nil {
		return nil, err
	}

	// Construct the node

	nodeOptions := &core.BuildCfg{
		Online:  true,
		Routing: libp2p.DHTOption, // This option sets the node to be a full DHT node (both fetching and storing DHT Records)
		// Routing: libp2p.DHTClientOption, // This option sets the node to be a client DHT node (only fetching records)
		Repo: repo,
		ExtraOpts: map[string]bool{
			"pubsub": true,
		},
	}

	node, err := core.NewNode(ctx, nodeOptions)
	if err != nil {
		return nil, err
	}

	// Attach the Core API to the constructed node
	return coreapi.NewCoreAPI(node)
}

// Spawns a node on the default repo location, if the repo exists
func spawnDefault(ctx context.Context) (icore.CoreAPI, error) {
	defaultPath, err := config.PathRoot()
	if err != nil {
		// shouldn't be possible
		return nil, err
	}

	if err := setupPlugins(defaultPath); err != nil {
		return nil, err
	}

	return createNode(ctx, defaultPath)
}

// Spawns a node to be used just for this run (i.e. creates a tmp repo)
func spawnEphemeral(ctx context.Context, repoPath *string) (icore.CoreAPI, error) {
	if err := setupPlugins(""); err != nil {
		return nil, err
	}

	// Create a Temporary Repo
	_, _ = createTempRepo(repoPath)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to create temp repo: %s", err)
	//}

	// Spawning an ephemeral IPFS node
	return createNode(ctx, *repoPath)
}

//

func connectToPeers(ctx context.Context, ipfs icore.CoreAPI, peers []string) error {
	var wg sync.WaitGroup
	peerInfos := make(map[peer.ID]*peer.AddrInfo, len(peers))
	for _, addrStr := range peers {
		addr, err := ma.NewMultiaddr(addrStr)
		if err != nil {
			return err
		}
		addrInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			return err
		}
		pi, ok := peerInfos[addrInfo.ID]
		if !ok {
			pi = &peer.AddrInfo{ID: addrInfo.ID}
			peerInfos[pi.ID] = pi
		}
		pi.Addrs = append(pi.Addrs, addrInfo.Addrs...)
	}

	wg.Add(len(peerInfos))
	for _, peerInfo := range peerInfos {
		go func(peerInfo *peer.AddrInfo) {
			defer wg.Done()
			_ = ipfs.Swarm().Connect(ctx, *peerInfo)
			//if err != nil {
			//	log.Printf("failed to connect to %s: %s", peerInfo.ID, err)
			//}
		}(peerInfo)
	}
	wg.Wait()
	return nil
}

func getUnixfsNode(path string) (files.Node, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	f, err := files.NewSerialFile(path, false, st)
	if err != nil {
		return nil, err
	}

	return f, nil
}

var bootstrapNodes = []string{
	// IPFS Bootstrapper nodes.
	"/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
	"/dnsaddr/bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa",
	"/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
	"/dnsaddr/bootstrap.libp2p.io/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
	"/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
	"/ip4/104.131.131.82/udp/4001/quic/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",

	// IPFS Cluster Pinning nodes
	"/ip4/138.201.67.219/tcp/4001/p2p/QmUd6zHcbkbcs7SMxwLs48qZVX3vpcM8errYS7xEczwRMA",
	"/ip4/138.201.67.219/udp/4001/quic/p2p/QmUd6zHcbkbcs7SMxwLs48qZVX3vpcM8errYS7xEczwRMA",
	"/ip4/138.201.67.220/tcp/4001/p2p/QmNSYxZAiJHeLdkBg38roksAR9So7Y5eojks1yjEcUtZ7i",
	"/ip4/138.201.67.220/udp/4001/quic/p2p/QmNSYxZAiJHeLdkBg38roksAR9So7Y5eojks1yjEcUtZ7i",
	"/ip4/138.201.68.74/tcp/4001/p2p/QmdnXwLrC8p1ueiq2Qya8joNvk3TVVDAut7PrikmZwubtR",
	"/ip4/138.201.68.74/udp/4001/quic/p2p/QmdnXwLrC8p1ueiq2Qya8joNvk3TVVDAut7PrikmZwubtR",
	"/ip4/94.130.135.167/tcp/4001/p2p/QmUEMvxS2e7iDrereVYc5SWPauXPyNwxcy9BXZrC1QTcHE",
	"/ip4/94.130.135.167/udp/4001/quic/p2p/QmUEMvxS2e7iDrereVYc5SWPauXPyNwxcy9BXZrC1QTcHE",

	// You can add more nodes here, for example, another IPFS node you might have running locally, mine was:
	"/ip4/127.0.0.1/tcp/4001/p2p/QmQQtheqZouh43hfV4E9woribXBGi6yLdefrrpvsCk7RxB",
	"/ip4/127.0.0.1/udp/4001/quic/p2p/QmQQtheqZouh43hfV4E9woribXBGi6yLdefrrpvsCk7RxB",
}

type Ipfs struct {
	api icore.CoreAPI
	ctx context.Context
}

func (this *Ipfs) Context() context.Context {
	return this.ctx
}

func NewIpfs(useLocalNode bool) (*Ipfs, error) {
	return NewIpfsWithRepo(useLocalNode, nil)
}

func NewIpfsWithRepo(useLocalNode bool, path *string) (*Ipfs, error) {
	var result Ipfs

	ctx, cancel := context.WithCancel(context.Background())
	runtime.SetFinalizer(&result, func(ipfs *Ipfs) {
		cancel()
	})
	result.ctx = ctx

	if useLocalNode {
		// Spawn a node using the default path (~/.ipfs), assuming that a repo exists there already
		ipfs, err := spawnDefault(ctx)
		if err != nil {
			return nil, err
		}
		result.api = ipfs
	} else {
		// Spawn a node using a temporary path, creating a temporary repo for the run
		ipfs, err := spawnEphemeral(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("failed to spawn ephemeral node: %s", err)
		}
		result.api = ipfs
	}

	err := connectToPeers(ctx, result.api, bootstrapNodes)
	if err != nil {
		return nil, err
	}

	//fmt.Println("IPFS node is running")

	return &result, nil
}

func (this *Ipfs) Add(node files.Node) (icorepath.Resolved, error) {
	var ipfs = this.api
	var ctx = this.ctx

	cidFile, err := ipfs.Unixfs().Add(ctx, node)
	if err != nil {
		return nil, fmt.Errorf("Could not add Node: %s", err)
	}

	return cidFile, nil
}

func (this *Ipfs) AddFromPath(path string) (icorepath.Resolved, error) {
	someFile, err := getUnixfsNode(path)
	if err != nil {
		return nil, fmt.Errorf("Could not create File Node: %s", err)
	}

	return this.Add(someFile)
}

func (this *Ipfs) AddFromReader(reader io.Reader) (icorepath.Resolved, error) {
	return this.Add(files.NewReaderFile(reader))
}

func (this *Ipfs) AddFromBytes(b []byte) (icorepath.Resolved, error) {
	return this.Add(files.NewBytesFile(b))
}

func (this *Ipfs) Cat(cidFile icorepath.Resolved) ([]byte, error) {
	var ipfs = this.api
	var ctx = this.ctx

	rootNodeFile, err := ipfs.Unixfs().Get(ctx, cidFile)
	if err != nil {
		return nil, fmt.Errorf("Could not get file with CID: %s", err)
	}

	switch rootNodeFile.(type) {
	case files.File:
	default:
		return nil, fmt.Errorf("%s is not a file", cidFile.String())
	}

	return ioutil.ReadAll(rootNodeFile.(files.File))
}

func (this *Ipfs) Ls(cidFile icorepath.Resolved) ([]files.Node, error) {
	var ipfs = this.api
	var ctx = this.ctx

	rootNodeDir, err := ipfs.Unixfs().Get(ctx, cidFile)
	if err != nil {
		return nil, fmt.Errorf("Could not get directory with CID: %s", err)
	}

	switch rootNodeDir.(type) {
	case files.Directory:
	default:
		return nil, fmt.Errorf("%s is not a file", cidFile.String())
	}

	result := make([]files.Node, 0)
	for entry := rootNodeDir.(files.Directory).Entries(); entry.Next(); {
		result = append(result, entry.Node())
	}
	return result, nil
}

func (this *Ipfs) Publish(topic string, toSend []byte) error {
	return this.api.PubSub().Publish(this.ctx, topic, toSend)
}

func (this *Ipfs) Subscribe(topic string, options ...options.PubSubSubscribeOption) (icore.PubSubSubscription, error) {
	return this.api.PubSub().Subscribe(this.ctx, topic, options...)
}

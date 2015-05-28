package commands

import (
	"fmt"
	"io"
	"strings"
	"time"

	mh "github.com/ipfs/go-ipfs/Godeps/_workspace/src/github.com/jbenet/go-multihash"
	context "github.com/ipfs/go-ipfs/Godeps/_workspace/src/golang.org/x/net/context"

	cmds "github.com/ipfs/go-ipfs/commands"
	u "github.com/ipfs/go-ipfs/util"
)

var PatchCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "Mutate a given merkledag object and return the modified node",
		ShortDescription: ``,
	},
	Options: []cmds.Option{},
	Arguments: []cmds.Argument{
		cmds.StringArg("root", true, false, "the hash of the node to modify"),
		cmds.StringArg("command", true, false, "the operation to perform"),
		cmds.StringArg("args", true, true, "extra arguments"),
	},
	Type: u.Key(""),
	Run: func(req cmds.Request, res cmds.Response) {
		nd, err := req.Context().GetNode()
		if err != nil {
			res.SetError(err, cmds.ErrNormal)
			return
		}

		rhash := u.B58KeyDecode(req.Arguments()[0])
		if rhash == "" {
			res.SetError(fmt.Errorf("incorrectly formatted root hash"), cmds.ErrNormal)
			return
		}

		ctx, _ := context.WithTimeout(req.Context().Context, time.Second*30)
		rnode, err := nd.DAG.Get(ctx, rhash)
		if err != nil {
			res.SetError(err, cmds.ErrNormal)
			return
		}

		switch req.Arguments()[1] {
		case "add-link":
			if len(req.Arguments()) < 4 {
				res.SetError(fmt.Errorf("not enough arguments for add-link"), cmds.ErrClient)
				return
			}

			hchild, err := mh.FromB58String(req.Arguments()[3])
			if err != nil {
				res.SetError(err, cmds.ErrNormal)
				return
			}

			k := u.Key(hchild)
			ctx, _ := context.WithTimeout(req.Context().Context, time.Second*30)
			childnd, err := nd.DAG.Get(ctx, k)
			if err != nil {
				res.SetError(err, cmds.ErrNormal)
				return
			}

			err = rnode.AddNodeLinkClean(req.Arguments()[2], childnd)
			if err != nil {
				res.SetError(err, cmds.ErrNormal)
				return
			}

			newkey, err := nd.DAG.Add(rnode)
			if err != nil {
				res.SetError(err, cmds.ErrNormal)
				return
			}

			res.SetOutput(newkey)

		case "rm-link":
			if len(req.Arguments()) < 3 {
				res.SetError(fmt.Errorf("not enough arguments for rm-link"), cmds.ErrClient)
				return
			}

			name := req.Arguments()[2]

			rnode.RemoveNodeLink(name)

			newkey, err := nd.DAG.Add(rnode)
			if err != nil {
				res.SetError(err, cmds.ErrNormal)
				return
			}

			res.SetOutput(newkey)
		default:
			res.SetError(fmt.Errorf("unrecognized subcommand"), cmds.ErrNormal)
			return
		}
	},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: func(res cmds.Response) (io.Reader, error) {
			k, ok := res.Output().(u.Key)
			if !ok {
				return nil, u.ErrCast()
			}

			return strings.NewReader(k.B58String() + "\n"), nil
		},
	},
}

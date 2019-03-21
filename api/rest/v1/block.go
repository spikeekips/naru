package restv1

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/nvellon/hal"

	sebakcommon "boscoin.io/sebak/lib/common"
	sebakresource "boscoin.io/sebak/lib/node/runner/api/resource"

	"github.com/spikeekips/naru/api/rest"
	"github.com/spikeekips/naru/newstorage/item"
)

func (h *Handler) GetBlock(w http.ResponseWriter, r *http.Request) {
	jw := rest.NewJSONWriter(w)

	vars := mux.Vars(r)
	hash := vars["hashOrHeight"]
	if hash == "" {
		jw.WriteObject(BadRequestParameter)
		return
	}

	var block item.Block
	var err error
	var height uint64
	height, err = strconv.ParseUint(hash, 10, 64)
	switch {
	case err == nil:
		block, err = item.GetBlockByHeight(h.st, height)
	case err != nil:
		block, err = item.GetBlock(h.st, hash)
	}
	if err != nil {
		jw.WriteObject(err)
		return
	}

	rs := sebakresource.NewBlock(&block.Block).Resource()

	rs.Links["self"] = hal.NewLink(strings.Replace(
		sebakresource.URLBlocks,
		"{id}",
		strconv.FormatUint(block.Height, 10),
		-1,
	))

	if block.Height != sebakcommon.GenesisBlockHeight {
		rs.AddLink("prev", hal.NewLink(strings.Replace(
			sebakresource.URLBlocks,
			"{id}",
			strconv.FormatUint(block.Height-1, 10),
			-1,
		)))
	}
	rs.AddLink("next", hal.NewLink(strings.Replace(
		sebakresource.URLBlocks,
		"{id}",
		strconv.FormatUint(block.Height+1, 10),
		-1,
	)))

	jw.WriteObject(rs)
}

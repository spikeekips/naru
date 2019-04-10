package restv1

import (
	"net/http"
	"strconv"
	"strings"

	sebakcommon "boscoin.io/sebak/lib/common"
	sebakresource "boscoin.io/sebak/lib/node/runner/api/resource"
	"github.com/gorilla/mux"
	"github.com/nvellon/hal"

	"github.com/spikeekips/naru/api/rest"
	resourcev1 "github.com/spikeekips/naru/api/rest/v1/resource"
	"github.com/spikeekips/naru/element"
)

func (h *Handler) GetBlock(w http.ResponseWriter, r *http.Request) {
	jw := rest.NewJSONWriter(w, r)

	vars := mux.Vars(r)
	hash := vars["hashOrHeight"]
	if hash == "" {
		jw.WriteObject(BadRequestParameter)
		return
	}

	var block element.Block
	var err error
	var height uint64
	height, err = strconv.ParseUint(hash, 10, 64)
	switch {
	case err == nil:
		block, err = h.potion.BlockByHeight(height)
	case err != nil:
		block, err = h.potion.Block(hash)
	}
	if err != nil {
		jw.WriteObject(err)
		return
	}

	rs := resourcev1.NewBlock(&block).Resource()

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

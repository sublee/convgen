//go:build convgen

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"example.com/convgenexample/api"
	"example.com/convgenexample/db"
	"example.com/convgenexample/pb"
	"github.com/sublee/convgen"
)

var mod = convgen.Module(
	convgen.RenameToLower(true, true),
	convgen.ForEnum(convgen.RenameTrimCommonPrefix(true, true)),
)

var (
	PBtoDB        = convgen.Struct[*pb.Job, db.Job](mod)
	PBtoDB_status = convgen.Enum[pb.Status, db.JobStatus](mod, db.JobStatusTodo, convgen.MatchSkip(pb.Status_STATUS_UNSPECIFIED, nil))
	DBtoPB        = convgen.Struct[db.Job, *pb.Job](mod)
	DBtoPB_status = convgen.Enum[db.JobStatus, pb.Status](mod, pb.Status_STATUS_UNSPECIFIED)
)

func Itoa64(i int64) string {
	return strconv.Itoa(int(i))
}

func Atoi64(s string) (int64, error) {
	i, err := strconv.Atoi(s)
	return int64(i), err
}

var (
	PBtoAPI = convgen.Struct[*pb.Job, api.Job](mod,
		convgen.MatchFunc(pb.Job{}.Id, api.Job{}.Id, Itoa64),
	)
	PBtoAPI_status = convgen.Enum[pb.Status, api.Status](mod, api.Unspecified)
	APItoPB        = convgen.StructErr[api.Job, *pb.Job](mod,
		convgen.MatchFuncErr(api.Job{}.Id, pb.Job{}.Id, Atoi64),
	)
	APItoPB_status = convgen.Enum[api.Status, pb.Status](mod, pb.Status_STATUS_UNSPECIFIED)
)

func pb2JSON(pb proto.Message) string {
	j, err := protojson.Marshal(pb)
	if err != nil {
		panic(err)
	}

	// Normalize the JSON output for consistent comparison.
	var b bytes.Buffer
	_ = json.Compact(&b, j)
	return b.String()
}

func main() {
	// Output: db.Job{ID:42, Status:"doing"}
	pbJob := &pb.Job{Id: 42, Status: pb.Status_STATUS_DOING}
	dbJob := PBtoDB(pbJob)
	fmt.Printf("%#v\n", dbJob)

	// Output: {"id":"42","status":"STATUS_DOING"}
	pbJob2 := DBtoPB(dbJob)
	pbJob2JSON := pb2JSON(pbJob2)
	fmt.Println(pbJob2JSON)

	// Output: api.Job{Id:"42", Status:"doing"}
	apiJob := PBtoAPI(pbJob)
	fmt.Printf("%#v\n", apiJob)

	// Output: {"id":"42","status":"STATUS_DOING"}
	pbJob3, err := APItoPB(apiJob)
	if err != nil {
		panic(err)
	}
	pbJob3JSON := pb2JSON(pbJob3)
	fmt.Println(pbJob3JSON)
}

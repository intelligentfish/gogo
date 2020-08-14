package robot

import (
	"fmt"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRobot_SendText(t *testing.T) {
	Convey("TestRobot_SendText", t, func() {
		hostname, err := os.Hostname()
		So(err, ShouldBeNil)
		content := fmt.Sprintf("%s/%d/%s", hostname, os.Getpid(), "TestRobot_SendText OK")
		result, err := Inst().SendText(content, nil, false)
		So(err, ShouldBeNil)
		So(result.ErrorCode, ShouldEqual, 0)
		So(result.ErrMsg, ShouldEqual, "ok")
	})
}

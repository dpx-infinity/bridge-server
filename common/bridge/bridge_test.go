/**
 * Date: 31.08.12
 * Time: 0:58
 *
 * @author Vladimir Matveev
 */
package bridge_test

import (
    "github.com/dpx-infinity/bridge-server/common/bridge"
    "github.com/dpx-infinity/bridge-server/common/conf"
    "github.com/dpx-infinity/bridge-server/common/msg"
    "github.com/dpx-infinity/bridge-server/common/plugins"
    . "launchpad.net/gocheck"
    "net"
    "testing"
    "time"
)

func Test(t *testing.T) {
    TestingT(t)
}

type BridgeSuite struct{}

var _ = Suite(BridgeSuite{})

func (_ BridgeSuite) TestLocalHandling(c *C) {
    addr, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:12345")
    cfg := &conf.Conf{
        Listeners: map[string]*conf.ListenerConf{
            "local": &conf.ListenerConf{
                Name: "local",
                Ports: []*conf.PortConf{
                    &conf.PortConf{
                        Type: conf.PortTypeTCP4,
                        Addr: addr,
                    },
                },
            },
        },
    }

    br := bridge.New(cfg)
    br.AddPlugin("echo-plugin", new(plugins.EchoPlugin))
    br.Start()
    time.Sleep(50 * time.Millisecond)

    cc, err := net.DialTCP("tcp4", nil, addr)
    c.Assert(err, IsNil)

    testEchoPlugin(cc, c)

    cc.Close()

    br.Stop()
}

func printCause(c *C, err error) {
    if err != nil {
        c.Log(err.(*msg.SerializeError).Cause())
    }
}

func testEchoPlugin(cc net.Conn, c *C) {
    bps := []byte{1, 2, 3, 4, 5}
    m := msg.CreateWithName("test-name")
    m.SetHeader("hdr1", "val1")
    m.SetHeader("hdr2", "val2")
    m.SetBodyPart("bp1", msg.BodyPartFromSlice(bps))

    err := msg.Serialize(m, cc)
    printCause(c, err)
    c.Assert(err, IsNil)

    rm, err := msg.DeserializeMessageName(cc)
    printCause(c, err)
    c.Assert(err, IsNil)
    c.Assert(rm.GetName(), Equals, m.GetName())

    err = msg.DeserializeMessageHeaders(cc, rm)
    printCause(c, err)
    c.Assert(err, IsNil)
    c.Assert(rm.GetHeader("hdr1"), Equals, m.GetHeader("hdr1"))
    c.Assert(rm.GetHeader("hdr2"), Equals, m.GetHeader("hdr2"))

    err = msg.DeserializeMessageBodyParts(cc, rm, msg.EmptyHook)
    c.Assert(err, IsNil)
    printCause(c, err)

    buf := make([]byte, rm.GetBodyPart("bp1").Size())
    n, err := rm.GetBodyPart("bp1").Read(buf)
    c.Assert(err, IsNil)
    c.Assert(n, Equals, len(bps))
    c.Assert(buf, DeepEquals, bps)
}

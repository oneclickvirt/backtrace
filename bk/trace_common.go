package backtrace

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/oneclickvirt/backtrace/model"
	. "github.com/oneclickvirt/defaultset"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// DefaultConfig is the default configuration for Tracer.
var DefaultConfig = Config{
	Delay:    50 * time.Millisecond,
	Timeout:  500 * time.Millisecond,
	MaxHops:  15,
	Count:    1,
	Networks: []string{"ip4:icmp", "ip4:ip", "ip6:ipv6-icmp", "ip6:ip"},
}

// DefaultTracer is a tracer with DefaultConfig.
var DefaultTracer = &Tracer{
	Config: DefaultConfig,
}

// Config is a configuration for Tracer.
type Config struct {
	Delay    time.Duration
	Timeout  time.Duration
	MaxHops  int
	Count    int
	Networks []string
	Addr     *net.IPAddr
}

// Tracer is a traceroute tool based on raw IP packets.
// It can handle multiple sessions simultaneously.
type Tracer struct {
	Config

	once     sync.Once
	conn     *net.IPConn      // Ipv4连接
	ipv6conn *ipv6.PacketConn // IPv6连接
	err      error

	mu   sync.RWMutex
	sess map[string][]*Session
	seq  uint32
}

// Trace starts sending IP packets increasing TTL until MaxHops and calls h for each reply.
func (t *Tracer) Trace(ctx context.Context, ip net.IP, h func(reply *Reply)) error {
	sess, err := t.NewSession(ip)
	if err != nil {
		return err
	}
	defer sess.Close()

	delay := time.NewTicker(t.Delay)
	defer delay.Stop()

	max := t.MaxHops
	for n := 0; n < t.Count; n++ {
		for ttl := 1; ttl <= t.MaxHops && ttl <= max; ttl++ {
			err = sess.Ping(ttl)
			if err != nil {
				return err
			}
			select {
			case <-delay.C:
			case r := <-sess.Receive():
				if max > r.Hops && ip.Equal(r.IP) {
					max = r.Hops
				}
				h(r)
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	if sess.isDone(max) {
		return nil
	}
	deadline := time.After(t.Timeout)
	for {
		select {
		case r := <-sess.Receive():
			if max > r.Hops && ip.Equal(r.IP) {
				max = r.Hops
			}
			h(r)
			if sess.isDone(max) {
				return nil
			}
		case <-deadline:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// NewSession returns new tracer session.
func (t *Tracer) NewSession(ip net.IP) (*Session, error) {
	t.once.Do(t.init)
	if t.err != nil {
		return nil, t.err
	}
	return newSession(t, shortIP(ip)), nil
}

func (t *Tracer) init() {
	// 初始化IPv4连接
	for _, network := range t.Networks {
		if strings.HasPrefix(network, "ip4") {
			t.conn, t.err = t.listen(network, t.Addr)
			if t.err == nil {
				go t.serve(t.conn)
				break
			}
		}
	}
	// 初始化IPv6连接
	for _, network := range t.Networks {
		if strings.HasPrefix(network, "ip6") {
			conn, err := net.ListenIP(network, t.Addr)
			if err == nil {
				t.ipv6conn = ipv6.NewPacketConn(conn)
				err = t.ipv6conn.SetControlMessage(ipv6.FlagHopLimit|ipv6.FlagSrc|ipv6.FlagDst|ipv6.FlagInterface, true)
				if err != nil {
					if model.EnableLoger {
						InitLogger()
						defer Logger.Sync()
						Logger.Info("设置IPv6控制消息失败: " + err.Error())
					}
					t.ipv6conn.Close()
					continue
				}
				go t.serveIPv6(t.ipv6conn)
				break
			}
		}
	}
}

// Close closes listening socket.
// Tracer can not be used after Close is called.
func (t *Tracer) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.conn != nil {
		t.conn.Close()
	}
	if t.ipv6conn != nil {
		t.ipv6conn.Close()
	}
}
func (t *Tracer) serve(conn *net.IPConn) error {
	defer conn.Close()
	buf := make([]byte, 1500)
	for {
		n, from, err := conn.ReadFromIP(buf)
		if err != nil {
			return err
		}
		err = t.serveData(from.IP, buf[:n])
		if err != nil {
			continue
		}
	}
}

func (t *Tracer) serveData(from net.IP, b []byte) error {
	if from.To4() == nil {
		// IPv6处理
		msg, err := icmp.ParseMessage(ProtocolIPv6ICMP, b)
		if err != nil {
			if model.EnableLoger {
				Logger.Warn("解析IPv6 ICMP消息失败: " + err.Error())
			}
			return err
		}
		// 记录所有收到的消息类型，帮助调试
		if model.EnableLoger {
			Logger.Info(fmt.Sprintf("收到IPv6 ICMP消息: 类型=%v, 代码=%v", msg.Type, msg.Code))
		}
		// 处理不同类型的ICMP消息
		switch msg.Type {
		case ipv6.ICMPTypeEchoReply:
			if echo, ok := msg.Body.(*icmp.Echo); ok {
				if model.EnableLoger {
					Logger.Info(fmt.Sprintf("处理IPv6回显应答: ID=%d, Seq=%d", echo.ID, echo.Seq))
				}
				return t.serveReply(from, &packet{from, uint16(echo.ID), 1, time.Now()})
			}
		case ipv6.ICMPTypeTimeExceeded:
			b = getReplyData(msg)
			if len(b) < ipv6.HeaderLen {
				if model.EnableLoger {
					Logger.Warn("IPv6时间超过消息太短")
				}
				return errMessageTooShort
			}
			// 解析原始IPv6包头
			if b[0]>>4 == ipv6.Version {
				ip, err := ipv6.ParseHeader(b)
				if err != nil {
					if model.EnableLoger {
						Logger.Warn("解析IPv6头部失败: " + err.Error())
					}
					return err
				}
				if model.EnableLoger {
					Logger.Info(fmt.Sprintf("处理IPv6时间超过: 目标=%v, FlowLabel=%d, HopLimit=%d",
						ip.Dst, ip.FlowLabel, ip.HopLimit))
				}
				return t.serveReply(ip.Dst, &packet{from, uint16(ip.FlowLabel), ip.HopLimit, time.Now()})
			}
		}
	} else {
		// 原有的IPv4处理逻辑
		msg, err := icmp.ParseMessage(ProtocolICMP, b)
		if err != nil {
			return err
		}
		if msg.Type == ipv4.ICMPTypeEchoReply {
			echo := msg.Body.(*icmp.Echo)
			return t.serveReply(from, &packet{from, uint16(echo.ID), 1, time.Now()})
		}
		b = getReplyData(msg)
		if len(b) < ipv4.HeaderLen {
			return errMessageTooShort
		}
		switch b[0] >> 4 {
		case ipv4.Version:
			ip, err := ipv4.ParseHeader(b)
			if err != nil {
				return err
			}
			return t.serveReply(ip.Dst, &packet{from, uint16(ip.ID), ip.TTL, time.Now()})
		default:
			return errUnsupportedProtocol
		}
	}
	return nil
}

func (t *Tracer) sendRequest(dst net.IP, ttl int) (*packet, error) {
	id := uint16(atomic.AddUint32(&t.seq, 1))
	var b []byte
	req := &packet{dst, id, ttl, time.Now()}
	if dst.To4() == nil {
		// IPv6
		b := newPacketV6(id, dst, ttl)
		if t.ipv6conn != nil {
			cm := &ipv6.ControlMessage{
				HopLimit: ttl,
			}
			_, err := t.ipv6conn.WriteTo(b, cm, &net.IPAddr{IP: dst})
			if err != nil {
				if model.EnableLoger {
					InitLogger()
					defer Logger.Sync()
					Logger.Info("发送IPv6请求失败: " + err.Error())
				}
				return nil, err
			}
			return req, nil
		}
		return nil, errors.New("IPv6连接不可用")
	} else {
		// IPv4
		b = newPacketV4(id, dst, ttl)
		_, err := t.conn.WriteToIP(b, &net.IPAddr{IP: dst})
		if err != nil {
			return nil, err
		}
		return req, nil
	}
}

func (t *Tracer) addSession(s *Session) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.sess == nil {
		t.sess = make(map[string][]*Session)
	}
	t.sess[string(s.ip)] = append(t.sess[string(s.ip)], s)
}

func (t *Tracer) removeSession(s *Session) {
	t.mu.Lock()
	defer t.mu.Unlock()
	a := t.sess[string(s.ip)]
	for i, it := range a {
		if it == s {
			t.sess[string(s.ip)] = append(a[:i], a[i+1:]...)
			return
		}
	}
}

func (t *Tracer) serveReply(dst net.IP, res *packet) error {
	if model.EnableLoger {
		Logger.Info(fmt.Sprintf("处理回复: 目标=%v, 来源=%v, ID=%d, TTL=%d",
			dst, res.IP, res.ID, res.TTL))
	}
	// 确保使用正确的IP格式进行查找
	shortDst := shortIP(dst)
	t.mu.RLock()
	defer t.mu.RUnlock()
	// // 调试输出会话信息
	// if model.EnableLoger && len(t.sess) > 0 {
	// 	for ip, sessions := range t.sess {
	// 		Logger.Info(fmt.Sprintf("会话信息: IP=%v, 会话数=%d",
	// 			net.IP([]byte(ip)), len(sessions)))
	// 	}
	// }
	// 查找对应的会话
	a := t.sess[string(shortDst)]
	if len(a) == 0 && model.EnableLoger {
		Logger.Warn(fmt.Sprintf("找不到目标IP=%v的会话", dst))
	}
	for _, s := range a {
		// if model.EnableLoger {
		// 	Logger.Info(fmt.Sprintf("处理会话响应: 会话目标=%v", s.ip))
		// }
		s.handle(res)
	}
	return nil
}

// Session is a tracer session.
type Session struct {
	t  *Tracer
	ip net.IP
	ch chan *Reply

	mu     sync.RWMutex
	probes []*packet
}

// NewSession returns new session.
func NewSession(ip net.IP) (*Session, error) {
	return DefaultTracer.NewSession(ip)
}

func newSession(t *Tracer, ip net.IP) *Session {
	s := &Session{
		t:  t,
		ip: ip,
		ch: make(chan *Reply, 64),
	}
	t.addSession(s)
	return s
}

// Ping sends single ICMP packet with specified TTL.
func (s *Session) Ping(ttl int) error {
	req, err := s.t.sendRequest(s.ip, ttl+1)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.probes = append(s.probes, req)
	s.mu.Unlock()
	return nil
}

// Receive returns channel to receive ICMP replies.
func (s *Session) Receive() <-chan *Reply {
	return s.ch
}

// isDone returns true if session does not have unresponsed requests with TTL <= ttl.
func (s *Session) isDone(ttl int) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.probes {
		if r.TTL <= ttl {
			return false
		}
	}
	return true
}

func (s *Session) handle(res *packet) {
	now := res.Time
	n := 0
	var req *packet
	if model.EnableLoger {
		Logger.Info(fmt.Sprintf("处理会话响应: 会话目标=%v, 响应源=%v, ID=%d, TTL=%d",
			s.ip, res.IP, res.ID, res.TTL))
	}
	s.mu.Lock()
	// // 打印出所有待处理的探测包
	// if model.EnableLoger && len(s.probes) > 0 {
	// 	Logger.Info(fmt.Sprintf("当前会话有 %d 个待处理的探测包", len(s.probes)))
	// 	for i, probe := range s.probes {
	// 		Logger.Info(fmt.Sprintf("探测包 #%d: ID=%d, TTL=%d, 时间=%v",
	// 			i, probe.ID, probe.TTL, probe.Time))
	// 	}
	// }
	// 查找匹配的请求包
	for _, r := range s.probes {
		if now.Sub(r.Time) > s.t.Timeout {
			// if model.EnableLoger {
			// 	Logger.Info(fmt.Sprintf("探测包超时: ID=%d, TTL=%d", r.ID, r.TTL))
			// }
			continue
		}
		// 对于IPv6 松散匹配
		if r.ID == res.ID || res.IP.To4() == nil {
			// if model.EnableLoger {
			// 	Logger.Info(fmt.Sprintf("找到匹配的探测包: ID=%d, TTL=%d", r.ID, r.TTL))
			// }
			req = r
			continue
		}
		s.probes[n] = r
		n++
	}
	s.probes = s.probes[:n]
	s.mu.Unlock()
	if req == nil {
		// if model.EnableLoger {
		// 	Logger.Warn(fmt.Sprintf("未找到匹配的探测包: 响应ID=%d", res.ID))
		// }
		return
	}
	hops := req.TTL - res.TTL + 1
	if hops < 1 {
		hops = 1
	}
	if model.EnableLoger {
		Logger.Info(fmt.Sprintf("创建响应: IP=%v, RTT=%v, Hops=%d",
			res.IP, res.Time.Sub(req.Time), hops))
	}
	select {
	case s.ch <- &Reply{
		IP:   res.IP,
		RTT:  res.Time.Sub(req.Time),
		Hops: hops,
	}:
	default:
		if model.EnableLoger {
			Logger.Warn("发送响应到通道失败，通道已满")
		}
	}
}

// Close closes tracer session.
func (s *Session) Close() {
	s.t.removeSession(s)
}

type packet struct {
	IP   net.IP
	ID   uint16
	TTL  int
	Time time.Time
}

func shortIP(ip net.IP) net.IP {
	if v := ip.To4(); v != nil {
		return v
	}
	return ip
}

func getReplyData(msg *icmp.Message) []byte {
	switch b := msg.Body.(type) {
	case *icmp.TimeExceeded:
		return b.Data
	case *icmp.DstUnreach:
		return b.Data
	case *icmp.ParamProb:
		return b.Data
	}
	return nil
}

var (
	errMessageTooShort     = errors.New("message too short")
	errUnsupportedProtocol = errors.New("unsupported protocol")
	errNoReplyData         = errors.New("no reply data")
)

// IANA Assigned Internet Protocol Numbers
const (
	ProtocolICMP     = 1
	ProtocolTCP      = 6
	ProtocolUDP      = 17
	ProtocolIPv6ICMP = 58
)

// Reply is a reply packet.
type Reply struct {
	IP   net.IP
	RTT  time.Duration
	Hops int
}

// Node is a detected network node.
type Node struct {
	IP  net.IP
	RTT []time.Duration
}

// Hop is a set of detected nodes.
type Hop struct {
	Nodes    []*Node
	Distance int
}

// Add adds node from r.
func (h *Hop) Add(r *Reply) *Node {
	var node *Node
	for _, it := range h.Nodes {
		if it.IP.Equal(r.IP) {
			node = it
			break
		}
	}
	if node == nil {
		node = &Node{IP: r.IP}
		h.Nodes = append(h.Nodes, node)
	}
	node.RTT = append(node.RTT, r.RTT)
	return node
}

// Trace is a simple traceroute tool using DefaultTracer.
func Trace(ip net.IP) ([]*Hop, error) {
	hops := make([]*Hop, 0, DefaultTracer.MaxHops)
	touch := func(dist int) *Hop {
		for _, h := range hops {
			if h.Distance == dist {
				return h
			}
		}
		h := &Hop{Distance: dist}
		hops = append(hops, h)
		return h
	}
	err := DefaultTracer.Trace(context.Background(), ip, func(r *Reply) {
		touch(r.Hops).Add(r)
	})
	if err != nil && err != context.DeadlineExceeded {
		return nil, err
	}
	sort.Slice(hops, func(i, j int) bool {
		return hops[i].Distance < hops[j].Distance
	})
	last := len(hops) - 1
	for i := last; i >= 0; i-- {
		h := hops[i]
		if len(h.Nodes) == 1 && ip.Equal(h.Nodes[0].IP) {
			continue
		}
		if i == last {
			break
		}
		i++
		node := hops[i].Nodes[0]
		i++
		for _, it := range hops[i:] {
			node.RTT = append(node.RTT, it.Nodes[0].RTT...)
		}
		hops = hops[:i]
		break
	}
	return hops, nil
}

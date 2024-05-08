package iamqp

import "github.com/streadway/amqp"

// ExchangeDeclare see amqp.ExchangeDeclare
func (s *Session) ExchangeDeclare(exchange, exchangeType string, opt ...ExDeclOpt) error {
	ch, err := s.conn.Channel()
	defer func() { _ = ch.Close() }()
	if err != nil {
		return err
	}
	// default option
	option := &ExchangeOption{
		durable:    true,
		autoDelete: false,
		internal:   false,
		passive:    false,
		noWait:     false,
		args:       nil,
	}
	for _, o := range opt {
		o(option)
	}
	if option.passive {
		err = ch.ExchangeDeclarePassive(exchange, exchangeType, option.durable, option.autoDelete,
			option.internal, option.noWait, option.args)
	} else {
		err = ch.ExchangeDeclare(exchange, exchangeType, option.durable, option.autoDelete,
			option.internal, option.noWait, option.args)
	}
	return err
}

// ExchangeOption ...
type ExchangeOption struct {
	durable    bool
	autoDelete bool
	internal   bool
	passive    bool
	noWait     bool
	args       amqp.Table
}

// ExDeclOpt ...
type ExDeclOpt func(option *ExchangeOption)

// ExchangeDurable ...
func ExchangeDurable(durable bool) ExDeclOpt {
	return func(option *ExchangeOption) {
		option.durable = durable
	}
}

// ExchangeAutoDelete ...
func ExchangeAutoDelete(autoDelete bool) ExDeclOpt {
	return func(option *ExchangeOption) {
		option.autoDelete = autoDelete
	}
}

// ExchangeInternal ...
func ExchangeInternal(internal bool) ExDeclOpt {
	return func(option *ExchangeOption) {
		option.internal = internal
	}
}

// ExchangePassive ...
func ExchangePassive(passive bool) ExDeclOpt {
	return func(option *ExchangeOption) {
		option.passive = passive
	}
}

// ExchangeNoWait ...
func ExchangeNoWait(noWait bool) ExDeclOpt {
	return func(option *ExchangeOption) {
		option.noWait = noWait
	}
}

// ExchangeArgs ...
func ExchangeArgs(args amqp.Table) ExDeclOpt {
	return func(option *ExchangeOption) {
		option.args = args
	}
}

// ExchangeType
const (
	ExchangeTypeDirect  = amqp.ExchangeDirect
	ExchangeTypeFanout  = amqp.ExchangeFanout
	ExchangeTypeTopic   = amqp.ExchangeTopic
	ExchangeTypeHeaders = amqp.ExchangeHeaders
)

// ExchangeDelete see amqp.channel.ExchangeDelete
func (s *Session) ExchangeDelete(name string, ifUnused, noWait bool) error {
	ch, err := s.conn.Channel()
	defer func() { _ = ch.Close() }()
	if err != nil {
		return err
	}
	return ch.ExchangeDelete(name, ifUnused, noWait)
}

// ExchangeBind see amqp.channel.ExchangeBind
func (s *Session) ExchangeBind(destination, key, source string, opt ...ExBindOption) error {
	ch, err := s.conn.Channel()
	defer func() { _ = ch.Close() }()
	if err != nil {
		return err
	}
	// default option
	option := &ExchangeBindOption{
		noWait: false,
		args:   nil,
	}
	for _, o := range opt {
		o(option)
	}
	return ch.ExchangeBind(destination, key, source, option.noWait, option.args)
}

// ExchangeBindOption ...
type ExchangeBindOption struct {
	noWait bool
	args   amqp.Table
}

// ExBindOption ...
type ExBindOption func(option *ExchangeBindOption)

// ExchangeBindNoWait ...
func ExchangeBindNoWait(noWait bool) ExBindOption {
	return func(option *ExchangeBindOption) {
		option.noWait = noWait
	}
}

// ExchangeBindArgs ...
func ExchangeBindArgs(args amqp.Table) ExBindOption {
	return func(option *ExchangeBindOption) {
		option.args = args
	}
}

// ExchangeUnbind see amqp.ExchangeUnbind
func (s *Session) ExchangeUnbind(destination, key, source string, opt ...ExUnbindOption) error {
	ch, err := s.conn.Channel()
	defer func() { _ = ch.Close() }()
	if err != nil {
		return err
	}
	// default option
	option := &ExchangeUnBindOption{
		noWait: false,
		args:   nil,
	}
	for _, o := range opt {
		o(option)
	}
	return ch.ExchangeUnbind(destination, key, source, option.noWait, option.args)
}

// ExchangeUnBindOption ...
type ExchangeUnBindOption struct {
	noWait bool
	args   amqp.Table
}

// ExUnbindOption ...
type ExUnbindOption func(option *ExchangeUnBindOption)

// ExchangeUnbindNoWait ...
func ExchangeUnbindNoWait(noWait bool) ExUnbindOption {
	return func(option *ExchangeUnBindOption) {
		option.noWait = noWait
	}
}

// ExchangeUnbindArgs ...
func ExchangeUnbindArgs(args amqp.Table) ExUnbindOption {
	return func(option *ExchangeUnBindOption) {
		option.args = args
	}
}

// QueueDeclare see amqp.QueueDeclare
func (s *Session) QueueDeclare(name string, opt ...QueueDeclOption) (amqp.Queue, error) {
	ch, err := s.conn.Channel()
	defer func() { _ = ch.Close() }()
	if err != nil {
		return amqp.Queue{}, err
	}
	// default option
	option := &QueueDeclareOption{
		durable:    true,
		autoDelete: false,
		noWait:     false,
		args:       nil,
	}
	for _, o := range opt {
		o(option)
	}
	if option.passive {
		return ch.QueueDeclarePassive(name, option.durable, option.autoDelete, option.exclusive, option.noWait, option.args)
	}
	return ch.QueueDeclare(name, option.durable, option.autoDelete, option.exclusive, option.noWait, option.args)
}

// QueueDeclareOption ...
type QueueDeclareOption struct {
	durable    bool
	autoDelete bool
	exclusive  bool
	noWait     bool
	passive    bool
	args       amqp.Table
}

// QueueDeclOption ...
type QueueDeclOption func(option *QueueDeclareOption)

// QueueDurable ...
func QueueDurable(durable bool) QueueDeclOption {
	return func(option *QueueDeclareOption) {
		option.durable = durable
	}
}

// QueueAutoDelete ...
func QueueAutoDelete(autoDelete bool) QueueDeclOption {
	return func(option *QueueDeclareOption) {
		option.autoDelete = autoDelete
	}
}

// QueueExclusive ...
func QueueExclusive(exclusive bool) QueueDeclOption {
	return func(option *QueueDeclareOption) {
		option.exclusive = exclusive
	}
}

// QueueNoWait ...
func QueueNoWait(noWait bool) QueueDeclOption {
	return func(option *QueueDeclareOption) {
		option.noWait = noWait
	}
}

// QueuePassive ...
func QueuePassive(passive bool) QueueDeclOption {
	return func(option *QueueDeclareOption) {
		option.passive = passive
	}
}

// QueueArgs ...
func QueueArgs(args amqp.Table) QueueDeclOption {
	return func(option *QueueDeclareOption) {
		option.args = args
	}
}

// QueueDelete see amqp.QueueDelete
func (s *Session) QueueDelete(name string, opt ...QueueDelOption) (int, error) {
	ch, err := s.conn.Channel()
	defer func() { _ = ch.Close() }()
	if err != nil {
		return 0, err
	}
	// default option
	option := &QueueDeleteOption{
		noWait: false,
	}
	for _, o := range opt {
		o(option)
	}
	return ch.QueueDelete(name, option.ifUnused, option.ifEmpty, option.noWait)
}

// QueueDeleteOption ...
type QueueDeleteOption struct {
	ifUnused bool
	ifEmpty  bool
	noWait   bool
}

// QueueDelOption ...
type QueueDelOption func(option *QueueDeleteOption)

// QueueDeleteIfUnused ...
func QueueDeleteIfUnused(ifUnused bool) QueueDelOption {
	return func(option *QueueDeleteOption) {
		option.ifUnused = ifUnused
	}
}

// QueueDeleteIfEmpty ...
func QueueDeleteIfEmpty(ifEmpty bool) QueueDelOption {
	return func(option *QueueDeleteOption) {
		option.ifEmpty = ifEmpty
	}
}

// QueueDeleteNoWait ...
func QueueDeleteNoWait(noWait bool) QueueDelOption {
	return func(option *QueueDeleteOption) {
		option.noWait = noWait
	}
}

// QueueBind see amqp.QueueBind
func (s *Session) QueueBind(name, key, exchange string, opt ...QBindOption) error {
	ch, err := s.conn.Channel()
	defer func() { _ = ch.Close() }()
	if err != nil {
		return err
	}
	// default option
	option := &QueueBindOption{
		noWait: false,
	}
	for _, o := range opt {
		o(option)
	}
	return ch.QueueBind(name, key, exchange, option.noWait, option.args)
}

// QueueBindOption ...
type QueueBindOption struct {
	noWait bool
	args   amqp.Table
}

// QBindOption ...
type QBindOption func(option *QueueBindOption)

// QueueBindNoWait ...
func QueueBindNoWait(noWait bool) QBindOption {
	return func(option *QueueBindOption) {
		option.noWait = noWait
	}
}

// QueueBindArgs ...
func QueueBindArgs(args amqp.Table) QBindOption {
	return func(option *QueueBindOption) {
		option.args = args
	}
}

// QueueUnbind see amqp.QueueUnbind
func (s *Session) QueueUnbind(name, key, exchange string, args amqp.Table) error {
	ch, err := s.conn.Channel()
	defer func() { _ = ch.Close() }()
	if err != nil {
		return err
	}
	return ch.QueueUnbind(name, key, exchange, args)
}

// QueueInspect see amqp.QueueInspect
func (s *Session) QueueInspect(name string) (amqp.Queue, error) {
	ch, err := s.conn.Channel()
	defer func() { _ = ch.Close() }()
	if err != nil {
		return amqp.Queue{}, err
	}
	return ch.QueueInspect(name)
}

// QueuePurge see amqp.QueuePurge
func (s *Session) QueuePurge(name string, noWait bool) (int, error) {
	ch, err := s.conn.Channel()
	defer func() { _ = ch.Close() }()
	if err != nil {
		return 0, err
	}
	return ch.QueuePurge(name, noWait)
}

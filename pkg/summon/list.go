package summon

// List list the content of the data tree.
func (s *Summoner) List(opts ...Option) ([]string, error) {
	s.Configure(opts...)

	return s.box.List(), nil
}

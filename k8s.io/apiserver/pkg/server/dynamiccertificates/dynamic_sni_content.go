package dynamiccertificates

type DynamicFileSNIContent struct {
	*DynamicCertKeyPairContent
	sniNames []string
}

var _ SNICertKeyContentProvider = &DynamicFileSNIContent{}
var _ ControllerRunner = &DynamicFileSNIContent{}

func NewDynamicSNIContentFromFiles(purpose, certFile, keyFile string, sniNames ...string) (*DynamicFileSNIContent, error) {
	servingContent, err := NewDynamicServingContentFromFiles(purpose, certFile, keyFile)
	if err != nil {
		return nil, err
	}

	ret := &DynamicFileSNIContent{
		DynamicCertKeyPairContent: servingContent,
		sniNames:                  sniNames,
	}
	if err := ret.loadCertKeyPair(); err != nil {
		return nil, err
	}

	return ret, nil
}

func (c *DynamicFileSNIContent) SNINames() []string {
	return c.sniNames
}

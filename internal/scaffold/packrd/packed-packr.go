// +build !skippackr
// Code generated by github.com/gobuffalo/packr/v2. DO NOT EDIT.

// You can use the "packr2 clean" command to clean up this,
// and any other packr generated files.
package packrd

import (
	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/packr/v2/file/resolver"
)

var _ = func() error {
	const gk = "143c4bbc43b44295b15363b75c6c44f9"
	g := packr.New(gk, "")
	hgr, err := resolver.NewHexGzip(map[string]string{
		"1704d77dd7a1222a503b8b8742eaebb5": "1f8b08000000000000ff54ceb16ac4300c80e1d97a0aa12986c3868e850e2d746c973e81e39313d373142c3b0d94be7b496eba51487cbfd610bfc3c458425e007259a5361cc0902801189a729bfbe8a2147f0d5bbeca96e3ecb597220b3dee27197b4ae126fe40abdf9e082c40ea4b3cf9c1e22f18eff14d768c616dbdb2629b1953beb1a2a47308aadc145b650633ca8ecf2f7882ee937f06fa3ad387411724e7fcfd9e2c1851f7bee736dcbf731f4753d4bdd6492f38ca6e2dfcc17f000000ffff5eccce02f1000000",
		"84e599561976e463821f7a70b5c31817": "1f8b08000000000000ffcacd4f29cd4955a8ae56d0f3cd4ff14bcc4d55a8ad05040000ffffa2e211e515000000",
		"9d793253ade4f41cd18decd837fbb2ae": "1f8b08000000000000ff8c514d8bdb30103d7b7ec5b0783f7cb04c7274096c68dc0d345facbb874221ab58922356968265fb62f4df8bec946c372decc596466fdebcf7265f66abd52c39489d1ca83d022ce7f97e37fffafd791aa4330c1fec912b8585a92aaa19c61d9e68f1564fbf4400f9cb7abddd64cffbcd7c9d613ac3be4792b7556534af37b4e2e8dc1fd47e91ed72bc103e6d2793c97abb78596533a3b134a8a46d301678dff7af7d8f61c864ed49c942d6cef57d4d75c9913c996f5271eb2b78aaa56e04dedcdae4d6fed237e72682fe956be6dcab73f74812420896b239b60752982a61b493cc74b238267690eb0111ccf33cfb91e33bdb426a86d45aded808802a9562f8f097ed2819f260c38fb378b891d2a04ff42316e05fd50be7105204c163f5e67dc427cf0281676ba562181b0c1faf158c1e4869003ea9ce8f1cdd462085665ce065efc3c0923718b7ef432bcda115822a337226dd743c4c816b26050405bb928677771f377dee01b25b6e373f532c14a71a866ffa5f8eb1e98c0dea0ae35af86c7e070000ffff15f984babe020000",
		"dbf7de2e0fc9ea50537bd015287f7aa8": "1f8b08000000000000ff74914f6fd34010c5effb29deb1961a47fcbb54e25028020e4508ca0121a49d7827f628f68eb53b761b45f9ee28b65ba814f634b2dfbcf9cd9bef43d769c40d19e16bd25102277ce35edddbff3de7ee1ac948dc6b16d3b447a3d93208bf0e0794b323a72fd4318ec7df178d599fafd6eb5aac193665a5dd3ad0284147a99a759ee4050219b97e21289dfba9032466a3b6859de63dfe0365f83383fc9573de7b57eb53df4975ab6111ac0f87673dc7e3a477d731c01a8e18329f8af3eee007ae06a34dcb30c58c0dca992d97ce2d39faac1dafa6af1eadec185917ae33a6f8ab9e51deed11784b436b978f13eea56de785a209d94c38b540e28c5bce520e6b8f2089abe928140312db9022c4327ab2a6c474b98a22360c1d3925091c9f8c284feb2c37292b8d5ba9cb3d75adc7565a86266cf618b2c47aee58a9c7b6a5ba74ee231b3a4d8c86db1ef762cdd920b16a7ce9dc8f3e909d6ce604ddeaf973ee45815bda313a0db2958a4c34665c500832959748dce9486d2efee53fe14f29b897053ec751770cdfd18ebd7b55e0bd769d18aa8662cdd9bd2e7047f5cc4ac8dc9d32ae3072caa211177de2ad3c2cbb8cbe706f0ad462e887dc60b532aa3334492d119f3e5cdfb83f010000ffff4e5f8b2a4d030000",
		"e3589c74ab00a3a43f6b4ed408f7d49f": "1f8b08000000000000ff24cb41aac4200c06e0bda7f869f795b7f536799ad64035a589656098bb0f8ceb8f6f855731888160d4ae9361a335edc8da7739b0cbc9618531a3ba5f96623cc4ebf8dfb2b658e891a28fe41a675b671b37b9680f0fdf26da13fe029d42c696f0fe041d7e0d2f72272cdb8c5c96c02fce3fff060000ffff9d5d3d4796000000",
	})
	if err != nil {
		panic(err)
	}
	g.DefaultResolver = hgr

	func() {
		b := packr.New("Summon scaffold template", "../templates/scaffold")
		b.SetResolver("Makefile", packr.Pointer{ForwardBox: gk, ForwardPath: "9d793253ade4f41cd18decd837fbb2ae"})
		b.SetResolver("README.md", packr.Pointer{ForwardBox: gk, ForwardPath: "dbf7de2e0fc9ea50537bd015287f7aa8"})
		b.SetResolver("assets/summon.config.yaml", packr.Pointer{ForwardBox: gk, ForwardPath: "e3589c74ab00a3a43f6b4ed408f7d49f"})
		b.SetResolver("go.mod", packr.Pointer{ForwardBox: gk, ForwardPath: "84e599561976e463821f7a70b5c31817"})
		b.SetResolver("{{.SummonerName}}/summon{{.go}}", packr.Pointer{ForwardBox: gk, ForwardPath: "1704d77dd7a1222a503b8b8742eaebb5"})
		}()

	return nil
}()

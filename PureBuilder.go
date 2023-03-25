package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	//"net/http"
	"net/url"
	"os"
	"strings"
	"strconv"
	"encoding/json"
	"net/http"
	"crypto/md5"
	"encoding/hex"
	"archive/zip"
	"path/filepath"
)

var pack Pack
var logger *log.Logger

type Pack struct {
	Name string `json:"name"`
	Version string `json:"version"`
	MinecraftVersion string `json:"mcv"`
	Mods []struct {
		Name string `json:"name"`
		Modtype string `json:"type"`
		Projectid string `json:"projectid"`
		Fileid string `json:"fileid, omitempty"`
		Side string `json:"side, omitempty"`
		Destination string `json:"destination, omitempty"`
	} `json:"mods"`
}

type ModrinthMod struct {
	Id string `json:"id"`
	Files []struct {
		Filename string `json:"filename"`
		Url string `json:"url"`
	} `json:"files"`
}

type CurseforgeMod struct {
	Data []struct {
		Id int `json:"id"`
		Filename string `json:"fileName"`
	} `json:"data"`
}

type GithubRelease struct {
	Assets string `json:"assets_url"`
}

type GithubAsset struct {
	Name string `json:"name"`
	Url string `json:"browser_download_url"`
}

type GithubMod struct{
	Filename string
	Url string
	MMCUrl string
	MD5 string
}

func main(){
	file, err := os.Create("purebuilder.log")
	eror(err)
	defer file.Close()

	logger = log.New(file, "purebuilder: ", 0)
	eror(os.RemoveAll("bld"))
	createdirs()
	jsonparse()
	download("https://maven.minecraftforge.net/net/minecraftforge/forge/1.7.10-10.13.4.1614-1.7.10/forge-1.7.10-10.13.4.1614-1.7.10-universal.jar", "tmp/forge-1.7.10-10.13.4.1614-1.7.10-universal.jar")
	copy("tmp/forge-1.7.10-10.13.4.1614-1.7.10-universal.jar", "bld/technic/bin/modpack.jar")
	downloadmcil()
	downloadunimixins()
	downloadlwjgl3ify()
	createinstance()
	copy("8.json", "bld/multimc/mmc-pack.json")
	createpackconfig()
	zipdirs()
}

func downloadmcil(){
	modrinthMod := apiModrinth("cUtsYbG5", pack.MinecraftVersion)
	download(modrinthMod[0].Files[0].Url, "tmp/"+modrinthMod[0].Files[0].Filename)
	copy("tmp/"+modrinthMod[0].Files[0].Filename, "bld/multimc/mods/"+modrinthMod[0].Files[0].Filename)
	copy("tmp/"+modrinthMod[0].Files[0].Filename, "bld/polymc/mods/"+modrinthMod[0].Files[0].Filename)
	copy("tmp/"+modrinthMod[0].Files[0].Filename, "bld/technic/mods/"+modrinthMod[0].Files[0].Filename)
	copy("tmp/"+modrinthMod[0].Files[0].Filename, "bld/modrinth/mods/"+modrinthMod[0].Files[0].Filename)
	copy("tmp/"+modrinthMod[0].Files[0].Filename, "bld/curse/mods/"+modrinthMod[0].Files[0].Filename)
}

func downloadunimixins(){
	modrinthMod := apiModrinth("ghjoiQAl", pack.MinecraftVersion)
	download(modrinthMod[0].Files[0].Url, "tmp/"+modrinthMod[0].Files[0].Filename)
	copy("tmp/"+modrinthMod[0].Files[0].Filename, "bld/multimc/mods/"+modrinthMod[0].Files[0].Filename)
	copy("tmp/"+modrinthMod[0].Files[0].Filename, "bld/polymc/mods/"+modrinthMod[0].Files[0].Filename)
	copy("tmp/"+modrinthMod[0].Files[0].Filename, "bld/technic/mods/"+modrinthMod[0].Files[0].Filename)
	copy("tmp/"+modrinthMod[0].Files[0].Filename, "bld/modrinth/mods/"+modrinthMod[0].Files[0].Filename)
	copy("tmp/"+modrinthMod[0].Files[0].Filename, "bld/curse/mods/"+modrinthMod[0].Files[0].Filename)
}

func downloadlwjgl3ify(){
	githubMod := apiGithub("GTNewHorizons/lwjgl3ify", false)
	download(githubMod.Url, "tmp/"+githubMod.Filename)
	download(githubMod.MMCUrl, "tmp/"+filenamefromurl(githubMod.MMCUrl))
	unzip("tmp/"+filenamefromurl(githubMod.MMCUrl), "bld/polymc/")
	copy("tmp/"+githubMod.Filename, "bld/polymc/mods/"+githubMod.Filename)
}

func createdirs(){
	eror(os.MkdirAll("bld/multimc/mods", os.ModePerm))
	eror(os.MkdirAll("bld/polymc/mods", os.ModePerm))
	eror(os.MkdirAll("bld/technic/bin", os.ModePerm))
	eror(os.MkdirAll("bld/technic/mods", os.ModePerm))
	eror(os.MkdirAll("bld/modrinth/mods", os.ModePerm))
	eror(os.MkdirAll("bld/curse/mods", os.ModePerm))
	//eror(os.MkdirAll("bld/generic", os.ModePerm))
	eror(os.MkdirAll("tmp", os.ModePerm))
	eror(os.MkdirAll("src/config", os.ModePerm))
	eror(os.MkdirAll("src/modpack", os.ModePerm))
	eror(os.MkdirAll("src/mods", os.ModePerm))
	eror(os.MkdirAll("out", os.ModePerm))
}

func zipdirs(){
	zipfile("bld/curse/", "out/curse.zip")
	zipfile("bld/modrinth/", "out/modrinth.zip")
	zipfile("bld/multimc/", "out/multimc.zip")
	zipfile("bld/polymc/", "out/polymc.zip")
	zipfile("bld/technic/", "out/technic.zip")
}

func createinstance(){
	f, err := os.Create("tmp/instance.cfg")
	eror(err)
	defer f.Close()
	writeline(f, "InstanceType=OneSix\n")
	writeline(f, "iconKey=flame\n")
	writeline(f, "name="+pack.Name+"\n")
	copy("tmp/instance.cfg", "bld/multimc/instance.cfg")
	copy("tmp/instance.cfg", "bld/polymc/instance.cfg")
}

func jsonparse(){
	modString, err := ioutil.ReadFile("pack.json")
	eror(err)
	eror(json.Unmarshal([]byte(modString), &pack));
}

func apiModrinth(projectid string, mcv string) []ModrinthMod {
	logger.Println("modrinth api request for projectid "+projectid+" and version "+mcv)
	body := request("https://api.modrinth.com/v2/project/"+projectid+"/version?game_versions=[%22"+mcv+"%22]")
	var modrinthMod []ModrinthMod
	eror(json.Unmarshal(body, &modrinthMod))
	return modrinthMod
}

func apiCurseforge(projectid string, mcv string) CurseforgeMod {
	logger.Println("curseforge api request for projectid "+projectid+" and version "+mcv)
	body := request("http://api-pocket.com/v1/mods/"+projectid+"/files?gameVersion="+mcv)
	var curseforgeMod CurseforgeMod
	eror(json.Unmarshal(body, &curseforgeMod))
	return curseforgeMod
}

func apiGithub(projectid string, hash bool) (GithubMod) {
	var githubmod GithubMod
	logger.Println("github api request for repoid "+projectid)
	body := request("https://api.github.com/repos/"+projectid+"/releases")
	var githubreleases []GithubRelease
	eror(json.Unmarshal(body, &githubreleases))
	body = request(githubreleases[0].Assets)
	var githubassets []GithubAsset
	eror(json.Unmarshal(body, &githubassets))
	for _, v := range githubassets {
		if (strings.Contains(strings.ToLower(v.Name), "dev") || strings.Contains(strings.ToLower(v.Name), "api") || strings.Contains(strings.ToLower(v.Name), "sources") || strings.Contains(strings.ToLower(v.Name), "patch") || strings.Contains(strings.ToLower(v.Name), "debug") || strings.Contains(strings.ToLower(v.Name), "agent")) {
			continue
		}
		if (strings.Contains(strings.ToLower(v.Name), "multimc")) {
			githubmod.MMCUrl = v.Url
			continue
		}
		Filename := filenamefromurl(v.Url)

		if(hash == true){
			download(v.Url, "tmp/"+Filename)
			githubmod.MD5 = md5file("tmp/"+Filename)
		}
		githubmod.Filename = Filename
		githubmod.Url = v.Url

	}
	return githubmod
}

func md5file(filepath string) (string) {
	file, err := os.Open(filepath)
	eror(err)
	defer file.Close()
	hash := md5.New()
	_, err = io.Copy(hash, file)
	eror(err)
	hashmd5 := hex.EncodeToString(hash.Sum(nil)[:16])
	logger.Println("file "+filepath+" hashed as "+hashmd5)
	return hashmd5
}

func request(s string) []byte{
	req, err := http.NewRequest("GET", s, nil)
	eror(err)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36")
	client := &http.Client{}
	res, err := client.Do(req)
	eror(err)
	body, err := ioutil.ReadAll(res.Body)
	eror(err)
	return body
}

func createpackconfig(){
	f, err := os.Create("pack.mcinstance")
	eror(err)
	defer f.Close()
	for i:=0; i < len(pack.Mods); i++ {
		if (i != 0){
			writeline(f, "\n")
		}
		writeline(f, "["+pack.Mods[i].Name+"]\n")

		if pack.Mods[i].Modtype == "modrinth" {
			writeline(f, "type = modrinth\n")
			modrinthMod := apiModrinth(pack.Mods[i].Projectid, pack.MinecraftVersion)
			writeline(f, "versionId = "+modrinthMod[0].Id+"\n")
			logger.Println("modrinth versionid determined for project "+pack.Mods[i].Projectid+" as "+modrinthMod[0].Id)
			if len(pack.Mods[i].Destination) > 0 {
				logger.Println("modrinth destination hard overwrote for project "+pack.Mods[i].Projectid+" to "+pack.Mods[i].Destination)
				writeline(f, "destination = "+pack.Mods[i].Destination+modrinthMod[0].Files[0].Filename+"\n")
			} else {
				writeline(f, "destination = mods/"+modrinthMod[0].Files[0].Filename+"\n")
			}
			writeline(f, "sourceFileName = "+modrinthMod[0].Files[0].Filename+"\n")
		}
		if pack.Mods[i].Modtype == "curseforge" {
			writeline(f, "type = curseforge\n")
			curseforgeMod := apiCurseforge(pack.Mods[i].Projectid, pack.MinecraftVersion)
			writeline(f, "projectId = "+pack.Mods[i].Projectid+"\n")
			if len(pack.Mods[i].Fileid) > 0 {
				logger.Println("curseforge fileid hard overwrote for project "+pack.Mods[i].Projectid+" to "+pack.Mods[i].Fileid)
				writeline(f, "fileId = "+pack.Mods[i].Fileid+"\n")
			} else {
				logger.Println("curseforge fileid determined for project "+pack.Mods[i].Projectid+" as "+strconv.Itoa(curseforgeMod.Data[0].Id))
				writeline(f, "fileId = "+strconv.Itoa(curseforgeMod.Data[0].Id)+"\n")
			}
			if len(pack.Mods[i].Destination) > 0 {
				logger.Println("curseforge destination hard overwrote for project "+pack.Mods[i].Projectid+" to "+pack.Mods[i].Destination)
				writeline(f, "destination = "+pack.Mods[i].Destination+curseforgeMod.Data[0].Filename+"\n")
			} else {
				writeline(f, "destination = mods/"+curseforgeMod.Data[0].Filename+"\n")
			}
		}
		if pack.Mods[i].Modtype == "github" {
			writeline(f, "type = url\n")
			githubMod := apiGithub(pack.Mods[i].Projectid, true)
			writeline(f, "url = "+githubMod.Url+"\n")
			if len(pack.Mods[i].Destination) > 0 {
				logger.Println("github destination hard overwrote for project "+pack.Mods[i].Projectid+" to "+pack.Mods[i].Destination)
				writeline(f, "destination = "+pack.Mods[i].Destination+githubMod.Filename+"\n")
			} else {
				writeline(f, "destination = mods/"+githubMod.Filename+"\n")
			}
			writeline(f, "MD5 = "+githubMod.MD5+"\n")
		}
		writeline(f, "side = "+pack.Mods[i].Side+"\n")
	}
}

func writeline(f *os.File, s string){
	_, err := f.WriteString(s)
	eror(err)
}

func eror(err error){
	if err != nil {
		fmt.Println(err)
	}
}

func download(fileURL string, location string){
	if fileexists(location){
		logger.Println(location+" exists not redownloading")
		return
	}
	fileName := filenamefromurl(fileURL)

	file, err := os.Create(location)
	eror(err)
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	resp, err := client.Get(fileURL)
	eror(err)
	defer resp.Body.Close()

	size, err := io.Copy(file, resp.Body)

	defer file.Close()

	logger.Println("downloaded file "+fileName+" to "+location+" with size "+strconv.Itoa(int(size)))
}

func zipfile(folder string, output string) {
	f, err := os.Create(output)
	eror(err)
	defer f.Close()
	writer := zip.NewWriter(f)
	defer writer.Close()
	eror(filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Method = zip.Deflate
		header.Name, err = filepath.Rel(filepath.Dir(folder), path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}
		headerWriter, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(headerWriter, f)
		return err
	}))
}

func unzip(src string, out string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(out, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(
				path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func filenamefromurl(furl string) (string) {
	fileURL, err := url.Parse(furl)
	eror(err)
	path := fileURL.Path
	segments := strings.Split(path, "/")
	fileName := segments[len(segments)-1]
	return fileName
}

func fileexists(path string) (bool) {
	_, err := os.Stat(path)
	if err == nil { return true }
	if os.IsNotExist(err) { return false }
	eror(err)
	return false
}

func copy(path string, out string){
	in, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Println(err)
			return
		}
	eror(ioutil.WriteFile(out, in, 0644))
}
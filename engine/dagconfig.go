package engine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"xxxx/dagengine/engine/graph"
	"xxxx/dagengine/engine/processor"

	"github.com/BurntSushi/toml"
)

// DAGConfig dag run config
type DAGConfig struct {
	opMeta []processor.OperatorMeta
	graph  graph.Cluster

	scriptPath string
}

func (p *DAGConfig) loadTomlScriptFile(tomlScript string) error {
	if _, err := toml.DecodeFile(tomlScript, &p.graph); err != nil {
		log.Printf("Failed to parse toml script file:%s with err:%v", tomlScript, err)
		return err
	}
	p.graph.Name = filepath.Base(tomlScript)
	err := p.graph.Build(p.opMeta)
	if nil != err {
		log.Printf("Failed to build graph with err:%v", err)
		return err
	}
	return nil
}

func (p *DAGConfig) loadTomlScriptContent(tomlScript string) error {
	if _, err := toml.Decode(tomlScript, &p.graph); err != nil {
		log.Printf("Failed to parse toml script file:%s with err:%v", tomlScript, err)
		return err
	}
	p.graph.Name = "DefaultCluster"
	err := p.graph.Build(p.opMeta)
	if nil != err {
		log.Printf("Failed to build graph with err:%v", err)
		return err
	}
	return nil
}

// DumpDot dump dot graph
func (p *DAGConfig) DumpDot() string {
	builder := &strings.Builder{}
	p.graph.DumpDot(builder)
	return builder.String()
}

// GenPng generate png from file
func (p *DAGConfig) GenPng(filePath string) error {
	if len(filePath) > 0 {
		p.scriptPath = filePath
	}
	dot := p.DumpDot()
	if strings.Contains(p.scriptPath, "&") ||
		strings.Contains(p.scriptPath, "|") ||
		strings.Contains(p.scriptPath, ";") {
		fmt.Println("command illegal")
		return fmt.Errorf("invalid file path:%v", p.scriptPath)
	}
	dotFile := p.scriptPath + ".dot"
	err := ioutil.WriteFile(dotFile, []byte(dot), 0600)
	if nil != err {
		log.Printf("Failed to write dot with err:%v", err)
		return err
	}
	pngFile := p.scriptPath + ".png"

	stdout, err := exec.Command("dot", "-Tpng", dotFile, "-o", pngFile).CombinedOutput()
	if err != nil {
		log.Printf(" exec cmd  failed with output:%s err:%v", stdout, err)
		return err
	}
	fmt.Printf("Write png into %s\n", pngFile)
	// fmt.Println(string(out[:]))
	return nil
}

// NewDAGConfigByFile new dag config by tmol
func NewDAGConfigByFile(opMetaFile string, tomlScript string) (*DAGConfig, error) {
	jsonFile, err := os.Open(opMetaFile)
	if err != nil {
		log.Printf("Failed to load op meta file:%s with err:%v", opMetaFile, err)
		return nil, err
	}
	config := &DAGConfig{}
	err = json.NewDecoder(jsonFile).Decode(&config.opMeta)
	if nil != err {
		log.Printf("Failed to parse op meta file:%s with err:%v", opMetaFile, err)
		return nil, err
	}
	err = config.loadTomlScriptFile(tomlScript)
	if nil != err {
		return nil, err
	}
	config.scriptPath = tomlScript
	return config, nil
}

// NewDAGConfigByContent new dag config by toml
func NewDAGConfigByContent(opMeta string, tomlScript string) (*DAGConfig, error) {
	opMeta = strings.TrimSpace(opMeta)
	config := &DAGConfig{}
	if len(opMeta) > 0 {
		err := json.Unmarshal([]byte(opMeta), &config.opMeta)
		if nil != err {
			log.Printf("Failed to parse op meta %s with err:%v", opMeta, err)
			return nil, err
		}
	}

	err := config.loadTomlScriptContent(tomlScript)
	if nil != err {
		return nil, err
	}
	config.scriptPath = tomlScript
	return config, nil
}

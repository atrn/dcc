// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//

package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const CompileCommandsFilename = "compile_commands.json"

type CompileCommand struct {
	Directory string `json:"directory"`
	Command   string `json:"command"`
	File      string `json:"file"`
}

func makeCompileCommands(sourceFilenames []string, compilerOptions *Options, objdir string) []CompileCommand {
	commands := make([]CompileCommand, len(sourceFilenames))
	for index, sourceFile := range sourceFilenames {
		command := fmt.Sprintf("%s %s -o %s -c %s", ActualCompiler.Name(), compilerOptions.String(), ObjectFilename(sourceFile, objdir), sourceFile)
		commands[index].Directory = DccCurrentDirectory
		commands[index].Command = command
		commands[index].File = sourceFile
	}
	return commands
}

func readCompileCommands(jsonFilename string) ([]CompileCommand, error) {
	file, err := os.Open(jsonFilename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	var commands []CompileCommand
	err = dec.Decode(&commands)
	if err != nil {
		return nil, err
	}
	return commands, nil
}

func writeCompileCommands(jsonFilename string, commands []CompileCommand) error {
	file, err := os.Create(jsonFilename)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	err = enc.Encode(commands)
	if err2 := file.Close(); err == nil {
		err = err2
	}
	return err
}

func WriteCompileCommandsDotJson(jsonFilename string, sourceFilenames []string, compilerOptions *Options, objdir string) error {
	return writeCompileCommands(jsonFilename, makeCompileCommands(sourceFilenames, compilerOptions, objdir))
}

func AppendCompileCommandsDotJson(jsonFilename string, sourceFilenames []string, compilerOptions *Options, objdir string) error {
	commands, err := readCompileCommands(jsonFilename)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	commands = append(commands, makeCompileCommands(sourceFilenames, compilerOptions, objdir)...)
	return writeCompileCommands(jsonFilename, commands)
}

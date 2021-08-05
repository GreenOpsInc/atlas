package com.greenops.workfloworchestrator.ingest.handling.testautomation;

import org.apache.logging.log4j.util.Strings;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

@Component
public class CommandBuilder {

    private List<String> commands;

    public CommandBuilder() {
        commands = new ArrayList<>();
    }

    public CommandBuilder createFile(String filename, String fileContents) {
        commands.add(Strings.join(List.of("echo", "'" + fileContents + "'", ">", filename), ' '));
        return this;
    }

    public CommandBuilder compile(String filename) {
        commands.add("chmod +x " + filename);
        return this;
    }

    public CommandBuilder ls() {
        commands.add("ls");
        return this;
    }

    public CommandBuilder cat(String filename) {
        commands.add("cat " + filename);
        return this;
    }

    public CommandBuilder executeFileContents(String fileContents) {
        commands.add(fileContents);
        return this;
    }

    public CommandBuilder executeExistingFile(String filename) {
        commands.add("./" + filename);
        return this;
    }

    public List<String> build() {
        return List.of(String.join("\n", commands));
    }
}

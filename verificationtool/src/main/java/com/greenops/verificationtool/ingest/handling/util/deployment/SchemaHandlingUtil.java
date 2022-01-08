package com.greenops.verificationtool.ingest.handling.util.deployment;

public class SchemaHandlingUtil {

    //Reference for escaping file contents: https://stackoverflow.com/questions/15783701/which-characters-need-to-be-escaped-when-using-bash
    public static String escapeFile(String fileContents) {
        var escapedFileContents = new StringBuilder();
        for (int i = 0; i < fileContents.length(); i++) {
            if (fileContents.charAt(i) == '\'') {
                escapedFileContents.append("'\\'");
            }
            escapedFileContents.append(fileContents.charAt(i));
        }
        return escapedFileContents.toString();
    }

    public static String getFileName(String filePathAndName) {
        var splitPath = filePathAndName.split("/");
        var idx = splitPath.length - 1;
        while (idx >= 0) {
            if (splitPath[idx].equals("")) {
                idx--;
            } else {
                break;
            }
        }
        if (idx >= 0) {
            return splitPath[idx];
        }
        return null;
    }

    public static String getFileNameWithoutExtension(String filePathAndName) {
        var filename = getFileName(filePathAndName);
        if (filename == null) return null;
        var idx = filename.length() - 1;
        while (idx >= 0) {
            if (filename.charAt(idx) == '.') {
                return filename.substring(0, idx);
            }
            idx--;
        }
        //Assuming its already an executable if there is no period
        return filename;
    }
}

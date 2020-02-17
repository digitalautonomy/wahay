#!/usr/bin/env ruby

$DOWNLOAD_DIR=ARGV[0]
$TMP_DIR=ARGV[1]

Dir.chdir($DOWNLOAD_DIR)

def generate_linux_packages_list(filename, isLatest)
    puts "<table class='table-bundled' cellpadding='0' cellspacing='0'>"
    File.read("#$TMP_DIR/supported-packages-distros.txt").each_line do |package_name|
        package_name.strip!
        if isLatest
            package = "wahay-#{package_name}-latest.deb"
            puts <<ENDOFHTML
  <tr>
    <td><a href="downloads/#{package}" title="#{package}">#{package_name}</a></td>
    <td><a href="downloads/#{package}.sha256sum">sha256sum</a></td>
    <td><a href="downloads/#{package}.sha256sum.asc">GPG signature</a></td>
  </tr>
ENDOFHTML
        else
            dir = Dir["linux-packages/#{filename}/*#{package_name}*.{deb,rpm}"]
            if dir.length > 0
                package = File.basename(dir.first)
                puts <<ENDOFHTML
                <tr>
                <td><a href="downloads/linux-packages/#{filename}/#{package}">#{package_name}</a></td>
                <td><a href="downloads/linux-packages/#{filename}/#{package}.sha256sum">sha256sum</a></td>
                <td><a href="downloads/linux-packages/#{filename}/#{package}.sha256sum.asc">GPG signature</a></td>
                </tr>
ENDOFHTML
            end
        end
    end
    puts "</table>"
end

def generate_bundle_list(filename, isLatest)
    puts "<table class='table-bundled' cellpadding='0' cellspacing='0'>"

    File.read("#$TMP_DIR/supported-bundle-distros.txt").each_line do |distro_name|
        distro_name.strip!

        if isLatest
            bundle = File.basename(Dir["*#{distro_name}*latest*.bz2"].first)
            puts <<ENDOFHTML
  <tr>
    <td><a href="downloads/#{bundle}">#{distro_name}</a></td>
    <td><a href="downloads/#{bundle}.sha256sum">sha256sum</a></td>
    <td><a href="downloads/#{bundle}.sha256sum.asc">GPG signature</a></td>
  </tr>
ENDOFHTML
        else
            # This will be of the form:
            # wahay-ubuntu-18_04-wahay-2020-02-13-500dfe5.tar.bz2
            bundle = File.basename(Dir["bundles/#{filename}/*#{distro_name}*.bz2"].first)

            puts <<ENDOFHTML
  <tr>
    <td><a href="downloads/bundles/#{filename}/#{bundle}">#{distro_name}</a></td>
    <td><a href="downloads/bundles/#{filename}/#{bundle}.sha256sum">sha256sum</a></td>
    <td><a href="downloads/bundles/#{filename}/#{bundle}.sha256sum.asc">GPG signature</a></td>
  </tr>
ENDOFHTML
        end
    end

    puts "</table>"
end

puts File.read(File.join(__dir__, "template_head.html"))

generate_bundle_list "wahay-latest", true

puts "</td>"
puts "<td valign='top'>"

generate_linux_packages_list "wahay-latest", true

puts "</td>"
puts "</tr>"

Dir["wahay*"].each do |filename|
    case filename
    when /sha256sum|wahay-latest|bz2|\.deb|\.rpm/
        # Ignore these files
    else
        puts <<ENDOFHTML
  <tr>
    <td><a href="downloads/#{filename}">#{filename}</a></td>
    <td><a href="downloads/#{filename}.sha256sum">sha256sum</a></td>
    <td><a href="downloads/#{filename}.sha256sum.asc">GPG signature</a></td>
    <td>
ENDOFHTML

        generate_bundle_list filename, false

        puts "</td>"
        puts "<td valign='top'>"
        generate_linux_packages_list filename, false
        puts "</td>"
        puts "</tr>"
    end
end

puts File.read(File.join(__dir__, "template_footer.html"))
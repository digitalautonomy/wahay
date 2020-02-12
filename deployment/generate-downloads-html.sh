#!/usr/bin/env bash

cd $1
TMP_DIR=$2

function generate_bundle_list()
{
local FILE=$1
local ISLATEST=$2

cat << end-of-html
<table class='table-bundled' cellpadding='0' cellspacing='0'>
end-of-html

while IFS= read -r DISTRO_NAME
do
if [ $ISLATEST == 0 ]
then
bundle=$(find . -name "*${DISTRO_NAME}*latest*.bz2" -exec basename {} \;)
cat << end-of-html
<tr>
<td><a href="downloads/${bundle}">${DISTRO_NAME}</a></td>
<td><a href="downloads/${bundle}.sha256sum">sha256sum</a></td>
<td><a href="downloads/${bundle}.sha256sum.asc">GPG signature</a></td>
</tr>
end-of-html
else
bundle=$(find bundles/${FILE}/ -name "*${DISTRO_NAME}*.bz2")
cat << end-of-html
<tr>
<td><a href="downloads/bundles/${bundle}">${DISTRO_NAME}</a></td>
<td><a href="downloads/bundles/${bundle}.sha256sum">sha256sum</a></td>
<td><a href="downloads/bundles/${bundle}.sha256sum.asc">GPG signature</a></td>
</tr>
end-of-html
fi
done < "$TMP_DIR"

cat << end-of-html
</table>
end-of-html

}

cat << end-of-html
<html>
<head>
<style>
html { padding:10px; width:95% }
.block-center { margin:auto; width:100%; }
.box { width:100%; border:1px solid #ADADAD; padding: 10px; margin:5px; }
#.box > p { font-weight:bold }
.table-files { width:100%; border:1px solid #333; border-collapse: collapse; }
.table-files tr th { border:1px solid #333; text-align: left; padding:5px; }
.table-files tr td { border:1px solid #333; padding:5px; }
.code{ background-color:#111;color:#FFF;font-family:Monospace;font-size:14px;margin:5px;padding:5px;width:90%; }
.table-bundled { width: 100%; }
.table-bundled tr td { border:1px solid #333 }
</style>
</head>
<body>
<div class="block-center">
<h1>Downloads</h1>
</div>
<div>
<p>Wahay is in an early stage of development. Not all of its functionalities have been developed and you may find errors! </p>
<p><strong>Download it at your own risk!</strong></p>
<p>At the moment it's available only for Linux 64 bits, in the future we plan to provide binaries for other operating systems. </p>
</div>

<div>
<h2>Linux downloads</h2>
<table class="table-files">
<tr>
<th>FILE</th>
<th>HASH</th>
<th>SIGNATURE</th>
<th style='text-align:center'>BUNDLED</th>
</tr>
<tr>
<td><a href="downloads/wahay-latest">wahay-latest</a></td>
<td><a href="downloads/wahay-latest.sha256sum">sha256sum</a></td>
<td><a href="downloads/wahay-latest.sha256sum">GPG signature</a></td>
<td>
end-of-html

generate_bundle_list "wahay-latest" 0

cat << end-of-html
</td>
</tr>
end-of-html

for filename in wahay*; do
    #Be sure not to list files with sha256sum string in their name
    ls $filename | grep 'sha256sum\|wahay-latest\|bz2' > /dev/null
        if [ $? -eq 1  ]
        then

cat << end-of-html
        <tr>
                <td><a href="downloads/$filename">$filename</a></td>
                <td><a href="downloads/$filename.sha256sum">sha256sum</a></td>
                <td><a href="downloads/$filename.sha256sum.asc">GPG signature</a></td>
                <td>
end-of-html

generate_bundle_list $filename 1

cat << end-of-html
                </td>
        </tr>
end-of-html

        fi
done
cat << end-of-html
</table>
</div>
<div class="box">
<p><strong>Hashes:</strong><br><br>
Digital hashes protects against unintentional modification of data in transit. They make sure that you get the same data as that sent from the website. They do not protect against any kind of attack.

<br>

To verify the integrity of the file you need to obtain the SHA256SUM of the binary you have downloaded and compare it to the correspondent wahay-xxxxx.sha256.sum. For example if you want to check the sha256sum of wahay-2020-01-22-9319d8d binary.

<br><br>
<div class="code">
$ sha256sum wahay-2020-01-22-9319d8d<br>
f75a4b04d05571d5eb7dff267c1efa996b1e24ff9a8d84c4fa1088141dc48cf8  wahay-2020-01-22-9319d8d<br>
</div>
<br>The output of the previous command should be compare with the content of wahay-2020-01-22-9319d8d.sha256sum file.
<br><br>

<code>
$ cat wahay-2020-01-22-9319d8d.sha256sum<br>
f75a4b04d05571d5eb7dff267c1efa996b1e24ff9a8d84c4fa1088141dc48cf8  bin/wahay-2020-01-22-9319d8d<br>
</code>
<br>

If the output of both is the same, then the binary has not been modified in transit, otherwise you have a corrupted file.

</div>
</p>
<div class="box">
<p><strong>Signatures:</strong><br><br>
Digital signatures ensures that what CAD intended to publish is the same as was published. It protects against attacks where the binary or source code has been modified by an attacker on the website, or modified in transit from the website to your system. It does NOT protect against attacks where the source code has been modified in our repositories, or when the build system has been compromised.<br><br>

1) Download and import CAD signing key (testing key at the moment):<br>
<code>
<br>$ wget https://staging.wahay.app/cad-testing-public-key.asc
</code>

<br><br>
2) Import the public key:<br>
<div class="code">
<br>$ gpg --import cad-testing-public-key.asc<br>
gpg: key A8854162D28F171E: public key "CAD Signing Key - testing (This is just a test key) <admin@autonomia.digital>" imported<br>
gpg: Total number processed: 1<br>
gpg:               imported: 1<br>
</div>
<br><br>

3) Verify the key<br>
<div class="code">
<br>$ gpg --verify wahay-2020-01-22-9319d8d.sha256sum.asc wahay-2020-01-22-9319d8d.sha256sum<br>
gpg: Signature made mié 22 ene 2020 10:06:02 -05<br>
gpg:                using EDDSA key A5DA0791073C1374BB2A98B3A5ABBD2E8E623464<br>
<strong>gpg: Good signature from "CAD Signing Key - testing (This is just a test key) \<admin@autonomia.digital\>"</strong> [unknown]<br>
<strong>gpg: WARNING: This key is not certified with a trusted signature!</strong><br>
gpg:          There is no indication that the signature belongs to the owner.<br>
Primary key fingerprint: 0124 2FFA B8CE 1EC0 C8F5  4456 A885 4162 D28F 171E<br>
     Subkey fingerprint: A5DA 0791 073C 1374 BB2A  98B3 A5AB BD2E 8E62 3464<br>
</div>
<br><br>
If you see the message: "gpg: Good signature from "CAD Signing Key - testing (This is just a test key) <admin@autonomia.digital>", that means that the signature is valid. However you would get the following warning: "This key is not certified with a trusted signature!". That is because the key is not trusted. At the moment don't trust in this key, when we have the final CAD signing key we would explain how to trust it.
</p>

</div>
</div>
end-of-html
echo "</body>"

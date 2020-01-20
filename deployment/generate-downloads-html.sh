#!/usr/bin/env bash

cd $1

cat << end-of-html
<html>
<head>
<style>
html { padding:10px; width:95% }
.block-center { margin:auto; width:100%; }
.box { width:100%; border:1px solid #ADADAD; padding: 10px; margin:5px; }
.box > p { font-weight:bold }
.table-files { width:100%; border:1px solid #333; border-collapse: collapse; }
.table-files tr th { border:1px solid #333; text-align: left; padding:5px; }
.table-files tr td { border:1px solid #333; padding:5px; }
</style>
</head>
<body>
<div class="block-center">
<h1>Downloads</h1>
</div>
<div>
<p>Tonio is in an early stage of development. Not all of its functionalities have been developed and you may find errors! </p>
<p><strong>Download it at your own risk!</strong></p>
<p>At the moment it's available only for Linux 64 bits, in the future we plan to provide binaries for other operating systems. </p>
</div>
<div class="box">
<p>Hashes:</p>
Digital hashes protects against unintentional modification of data in transit. They make sure that you get the same data as was sent from the website. They do not protect against any kind of attack.
</div>
<div class="box">
<p>Signatures:</p>
Digital signatures ensures that what CAD intended to publish is the same as was published. It protects against attacks where the binary or source code has been modified by an attacker on the website, or modified in transit from the website to your system. It does NOT protect against attacks where the source code has been modified in our repositories, or when the build system has been compromised.
</div>
</div>
<div>
<h2>Linux downloads</h2>
<table class="table-files">
<tr>
<th>FILE</th>
<th>HASH</th>
<th>SIGNATURE</th>
</tr>
<tr>
<td><a href="downloads/tonio-latest">tonio-latest</a></td>
<td><a href="downloads/tonio-latest.sha256sum">sha256sum</a></td>
<td><a href="downloads/tonio-latest.sha256sum">GPG signature</a></td>
</tr>

end-of-html


for filename in tonio*; do
    #Be sure not to list files with sha256sum string in their name
    ls $filename | grep  "sha256sum\|tonio-latest" > /dev/null
        if [ $? -eq 1  ]
        then

cat << end-of-html
        <tr>
                <td><a href="downloads/$filename">$filename</a></td>
                <td><a href="downloads/$filename.sha256sum">sha256sum</a></td>
                <td><a href="downloads/$filename.sha256sum.asc">GPG signature</a></td>
        </tr>
end-of-html
    
        fi
done
cat << end-of-html
</table>
</div>
end-of-html
echo "</body>"


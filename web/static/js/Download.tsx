import React from "react";
import { Link } from "react-router-dom";
import axios from "axios";
import marked from "marked";
import { toast } from "react-toastify";

import "../css/download.scss";
import useMemoryState from "./MemoryState";

interface GitHubRelease {
    name:           string;
    draft:          boolean;
    prerelease:     boolean;
    published_at:   Date;
    assets:         Asset[];
    body:           string;
}

interface Asset {
    name:                   string;
    size:                   number;
    download_count:         number;
    browser_download_url:   string;
}

export default function() {
    const [releases, setReleases] = useMemoryState("githubReleases", []) as [GitHubRelease[], Function];

    if (releases.length == 0) {
        axios.get("https://api.github.com/repos/lieuweberg/rich-destiny/releases", {
            timeout: 1000 * 5
        }).then(res => {
            setReleases(res.data);
        }).catch(err => {
            let errorMsg = [<p>Refresh the page or go directly to
            the <a href="https://github.com/lieuweberg/rich-destiny/releases/latest" target="_blank"
            rel="noopener noreferrer">releases page</a> to download rich-destiny.</p>];

            if (err.response) {
                errorMsg.push(<p>GitHub did not respond with a 2xx status code, got ${err.response.status}:
                <code>${err.response.data}</code></p>);
            } else if (err.request) {
                errorMsg.push(<p>GitHub could not be reached (no internet?).</p>);
            } else {
                errorMsg.push(<p>Could not make request to GitHub... <code>${err.message}</code></p>);
            }
            toast.error(<div>{errorMsg[1]}{errorMsg[0]}</div>)
        })
    }

    let latest: GitHubRelease;
    for (let r of releases) {
        if (!(r.prerelease || r.draft)) {
            latest = r;
            break;
        }
    }
    
    let version = "vX.Y.Z";
    let url = "https://github.com/lieuweberg/rich-destiny/releases/latest";
    let size = "??";
    // let downloads = ":(";
    if (latest) {
        let asset = latest.assets[0];
        version = latest.name;
        url = asset.browser_download_url;
        size = (asset.size / 1e6).toFixed(1).toString();
        // downloads = asset.download_count.toString();
    }

    return <>
        <div className="generic-text top-text">
            <h1>Download</h1>
            <p>You can download the latest (non-pre) release here. Source code (in case
            you want to view that) can be found in the <a href="https://github.com/lieuweberg/rich-destiny"
            target="_blank" rel="noopener noreferrer">GitHub</a> repo alongside old releases. Note
            that old releases don't do anything special, please download the most recent release.
            Clicking the download button here downloads the same file as GitHub releases provides.</p>
            <p>Installation instructions can be found below, in case they are needed. It is recommended
            to fully read these prior to installation.</p></div>
            <div className="generic-text no-padding">

            <a id="release-link" href={url} rel="noopener noreferrer">Download {version}</a>
            <p>Size: {size}MiB <br/>
            {/* Downloads: {downloads}</p> */}</p>
            <div dangerouslySetInnerHTML= {{ __html: marked(latest ? latest.body : "") }} />
        </div>
        <div className="generic-text">
            <h1>Installation</h1>
            <ol id="install-steps">
                <li>
                    <p>First download rich-destiny by clicking the big 'Download {version}' button above.
                    Download the file to where you want the program to reside. It will create additional
                    files in its folder. Recommended location: <code>C:\Users\YOURNAME\rich-destiny</code></p>
                </li>
                <li>
                    <p>Double click <code>rich-destiny.exe</code>. Windows Defender will pop up saying
                    that the file is from an unknown source and can't be trusted<sup>[1]</sup>.
                    Click 'More info' and then 'Run anyway'.</p>
                </li>
                <li>
                    <p>Now another window will pop up asking for administrator permissions. This is so
                    rich-destiny can install itself into the service manager, allowing it to run in
                    the background. Click 'Yes'.</p>
                </li>
                <li>
                    <p>Since there is no GUI, a command window will pop up. Take your time to read what
                    it says. A browser tab will shortly pop up to log in with Bungie.net.</p>
                    <p>That's it. You're done. Head to the <Link to="/cp">control panel</Link> to
                    configure rich-destiny! Enjoy :) If you have any questions, feel free to ask in
                    the <a href="https://discord.gg/UNU4UXp" target="_blank"
                    rel="noopener noreferrer">Discord</a> server.</p>
                </li>
            </ol>
            <p><small>[1]: Requires a code signing certificate which is expensive ($200-$500/year,
            depending on the vendor). Unsigned executables will build up trust over time, however, so at
            a certain point it will stop presenting the unsafe executable message for everyone.</small></p>
        </div>
    </>
}
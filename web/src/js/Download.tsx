import React, { useState } from "react";
import { Link } from "react-router-dom";
import axios from "axios";
import { marked } from "marked";
import { toast } from "react-toastify";

import "../css/download.scss";
import useMemoryState from "./MemoryState";

// // @ts-expect-error
// import decoration from "../images/s19-towerisland.webp";

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

interface ParsedRelease {
    version:    string;
    body:       string;
    url:        string;
    size:       string;
    downloads:  number;
    date:       string;
    year:       number;
}

export default function() {
    const [releases, setReleases] = useMemoryState("githubReleases", []) as [GitHubRelease[], Function];
    const [releaseIndex, setReleaseIndex] = useState(0);

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

    function parseRelease(githubRelease: GitHubRelease): ParsedRelease {
        let asset = githubRelease.assets[0];
        console.log(githubRelease);
        let d = (new Date(githubRelease.published_at));
        let r: ParsedRelease = {
            version: githubRelease.name,
            body: githubRelease.body,
            url: asset.browser_download_url,
            size: (asset.size / 1e6).toFixed(1),
            downloads: asset.download_count
                + (githubRelease.assets[1] ? githubRelease.assets[1].download_count : 0),
            date: d.toLocaleString("en-us", { day: "2-digit", month: "long" }),
            year: d.getFullYear()
        };

        return r;
    }

    let selectedReleases: ParsedRelease[] = [];
    for (let r of releases) {
        if (!(r.prerelease || r.draft)) {
            if (selectedReleases.length == 5) {
                break;
            } else {
                selectedReleases.push(parseRelease(r));
            }
        }
    }

    let r = selectedReleases[releaseIndex];
    if (!r) {
        r = {
            version: "vX.Y.Z",
            body: "Fetching releases...",
            url: "https://github.com/lieuweberg/rich-destiny/releases/latest",
            size: "??",
            downloads: Infinity,
            date: "Soon™",
            year: 2077
        }
    }
    
    return <>
        {/* <img id="dl-decoration" className="sidebar-decoration" src={decoration} alt="" /> */}

        <div className="generic-text">
            <h1>Download</h1>

            <p>You can download the latest release here. Source code can be
            found in the <a href="https://github.com/lieuweberg/rich-destiny" target="_blank"
            rel="noopener noreferrer">GitHub repo</a> alongside old releases.</p>
            <p>Installation instructions can be found below. It is recommended
            to fully read these prior to installation.</p>
            <p><b>By clicking the Download button below, you agree to
            the <a href="https://github.com/lieuweberg/rich-destiny/blob/master/LICENSE.md" target="_blank"
            rel="noopener noreferrer">Terms and Conditions</a></b> (don't worry, it's just the open source license),
            including section 15 (Disclaimer of Warranty) and 16 (Limitation of Liability).</p>

            <div id="download-box" className="boxed">
                <div>
                    <a className="button" href={r.url}>Download {r.version}</a>
                    <div dangerouslySetInnerHTML= {{ __html: marked(r ? r.body : "") }} />
                </div>
                <div>
                    <h2>About</h2>
                    <p>Version: {r.version}<br/>
                    Size: {r.size}MiB <br/>
                    Date: {r.date} {r.year} <br/>
                    Downloads: {r.downloads}</p>
                </div>
                <div>
                    <h2>History</h2>
                    <p>Click to view release notes.</p>
                    <ul id="old-releases">
                        {selectedReleases.map((release, i) => (
                            <li key={i}><a href="#" onClick={e => {
                                e.preventDefault();
                                setReleaseIndex(i);
                            }}>{release.version}</a> <span className="grey"> — {release.date}</span></li>
                        ))}
                    </ul>
                </div>
            </div>
        </div>
        <div className="generic-text">
            <h1>Installation</h1>
            <p>Don't worry, it's really simple.</p>
            <ol id="install-steps">
                <li>
                    <p>Download rich-destiny by clicking the 'Download {r.version}' button above.</p>
                </li>
                <li>
                    <p>Open the downloaded <code>rich-destiny.exe</code>. Windows SmartScreen will pop up saying
                    that the file is from an unknown source and can't be trusted<sup>[1]</sup>. <b>Click
                    'More info' and then 'Run anyway'.</b></p>
                </li>
                <li>
                    <p>A text window will pop up. It will ask where you want to install rich-destiny.
                    The location that is guaranteed to work is the default, so <b>type 'default'
                    and hit enter</b>. A browser tab
                    will shortly open to log in with Bungie.net.</p>
                    <p>That's it. You're done. Head to the <Link to="/cp">control panel</Link> to
                    configure rich-destiny! Enjoy :) If you have any questions, feel free to ask in
                    the <a href="https://richdestiny.app/discord" target="_blank"
                    rel="noopener noreferrer">Discord</a> server.</p>
                </li>
            </ol>
            <p><small>[1]: Requires a code signing certificate which is expensive ($200-$500/year,
            depending on the vendor). Unsigned executables will build up trust over time, however, so at
            a certain point it will stop presenting the unsafe executable message for everyone.</small></p>
        </div>
    </>
}
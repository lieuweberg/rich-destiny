import React from "react";
import { Link } from "react-router-dom";

import PresenceCard, { PresenceCardProps } from "./components/PresenceCard";

import examplePresencesJSON from "../misc/example-presences.json";

import "../css/home.scss";

export default function() {
    let examplePresences: PresenceCardProps[] = [];
    for (let i = 0; i < 3*6; i++) {
        let p = examplePresencesJSON[i];
        examplePresences.push({
            largeImage: p.largeImage,
            description: p.description,
            state: p.state,
            time: p.time
        })
    }

    return <>
        <div id="title">
            <h1>rich-destiny</h1>
            <p>a discord rich presence tool for destiny 2 (pc)</p>
            <div>
                <Link to="/download">Download</Link>
                <Link to="/cp">Control Panel</Link>
            </div>
        </div>
        <div id="sliders">
            <div className="animated-row-1">
                <div className="slider-item">
                    {examplePresences.slice(0,6).map((p, i) => <PresenceCard key={i} {...p} />)}
                </div>
                <div className="slider-item">
                    {examplePresences.slice(0,6).map((p, i) => <PresenceCard key={i} {...p} />)}
                </div>
            </div>
            <div className="animated-row-2">
                <div className="slider-item">
                    {examplePresences.slice(6,12).map((p, i) => <PresenceCard key={i} {...p} />)}
                </div>
                <div className="slider-item">
                    {examplePresences.slice(6,12).map((p, i) => <PresenceCard key={i} {...p} />)}
                </div>
            </div>
            <div className="animated-row-3">
                <div className="slider-item">
                    {examplePresences.slice(12).map((p, i) => <PresenceCard key={i} {...p} />)}
                </div>
                <div className="slider-item">
                    {examplePresences.slice(12).map((p, i) => <PresenceCard key={i} {...p} />)}
                </div>
            </div>
        </div>

        <div id="info-cards">
            <div className="card">
                <div className="image">
                    <img src="https://icongr.am/entypo/cog.svg?size=300&color=ffffff" alt="screenshot showing tooltip"/>
                </div>
                <div>
                    <h3>Background process</h3>
                 <p>Fully runs in the background and needs no further maintenance. Auto updates by default.</p>
                </div>
            </div>
            <div className="card">
                <img src="https://f.lieuweberg.com/HHKjbd.png" alt="screenshot showing tooltip"/>
                <div>
                    <h3>Class & power</h3>
                 <p>Displays current class and power level as a tooltip on the class icon.</p>
                </div>
            </div>
            <div className="card">
                <div className="image">
                    <img src="https://icongr.am/entypo/drive.svg?size=300&color=ffffff" alt="screenshot showing tooltip"/>
                </div>
                <div>
                    <h3>Small footprint</h3>
                 <p>Around 23MiB in size and uses close to no CPU and 4-7MB of memory. GUI is this website.</p>
                </div>
            </div>
        </div>

        <div className="generic-text">
            <h1>What's so special about this?</h1>
            <p>After <Link to="/download">downloading</Link>, this program installs itself into the
            Windows service manager. Every time you now start your computer, this program will also
            be started, do some minimal work and then sit idly, waiting until you start Destiny 2.</p>
            <p>Many Alternatives have a built-in GUI (Graphical User Interface). This adds a lot of
            overhead while rarely being used, mostly to disk size. That's why rich-destiny allows this
            site to communicate with itself, providing the GUI <i>through this site</i>. This has many
            advantages, mainly hugely reduced download size (and update size) and instant GUI updates.</p>
        </div>
    </>
}
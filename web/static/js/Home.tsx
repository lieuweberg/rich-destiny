import React from "react";
import { Link } from "react-router-dom";

import PresenceCard, { PresenceCardProps } from "./components/PresenceCard";

import "../css/home.scss";

// @ts-expect-error
import home1 from "../images/home1.png";
// @ts-expect-error
import home2 from "../images/home2.png";

export default function() {
    let presences: PresenceCardProps[] = [];
    for (let i = 0; i < 3*6; i++) {
        let p = examplePresences[i];
        presences.push({
            largeImage: p.i,
            description: p.d,
            state: p.s,
            initialTime: p.t
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
                    {presences.slice(0,6).map((p, i) => <PresenceCard key={i} {...p} />)}
                </div>
                <div className="slider-item">
                    {presences.slice(0,6).map((p, i) => <PresenceCard key={i} {...p} />)}
                </div>
            </div>
            <div className="animated-row-2">
                <div className="slider-item">
                    {presences.slice(6,12).map((p, i) => <PresenceCard key={i} {...p} />)}
                </div>
                <div className="slider-item">
                    {presences.slice(6,12).map((p, i) => <PresenceCard key={i} {...p} />)}
                </div>
            </div>
            <div className="animated-row-3">
                <div className="slider-item">
                    {presences.slice(12).map((p, i) => <PresenceCard key={i} {...p} />)}
                </div>
                <div className="slider-item">
                    {presences.slice(12).map((p, i) => <PresenceCard key={i} {...p} />)}
                </div>
            </div>
        </div>

        <div id="info-cards">
            <div className="card">
                <div className="image">
                    <img src={home1} alt="screenshot showing tooltip"/>
                </div>
                <hr/>
                <div>
                    <h3>Forge your fireteam</h3>
                    <p>No more sharing join codes: have anyone effortlessly join your fireteam by clicking a button in
                        your status. <small>(optional)</small></p>
                </div>
            </div>
            <div className="card">
                <div className="image">
                    <img src={home2} alt="screenshot showing tooltip"/>
                </div>
                <hr/>
                <div>
                    <h3>Make your class proud</h3>
                    <p>Displays current class and power level as a tooltip on the class icon.</p>
                </div>
            </div>
            <div className="card">
                <div className="image">
                    <img src="https://icongr.am/entypo/drive.svg?size=300&color=ffffff" alt="screenshot showing tooltip"/>
                </div>
                <hr/>
                <div>
                    <h3>It just works</h3>
                    <p>Runs in the background and is tiny in size. Auto updates by default. Fully self-contained.</p>
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

const examplePresences = [
    {i:"strikes",d:"Normal Strikes - Nessus",s:"The Insight Terminus",t:"08:42"},
    {i:"destinylogo",d:"In orbit",s:"space 🌌 (customisable)",t:"04:13"},
    {i:"storypvecoopheroic",d:"Story - Europa",s:"The New Kell",t:"13:24"},
    {i:"raid",d:"Raid - The Dreaming City",s:"Last Wish",t:"02:15:11"},
    {i:"control",d:"Control - The Crucible",s:"Pacifica",t:"05:58"},
    {i:"gambit",d:"Gambit",s:"Emerald Coast",t:"07:26"},
    {i:"crucible",d:"Mayhem - The Crucible",s:"Javelin-4",t:"02:55"},
    {i:"socialall",d:"Social - Earth",s:"Tower",t:"13:17"},
    {i:"raid",d:"Raid - Europa",s:"Deep Stone Crypt",t:"59:08"},
    {i:"strikes",d:"Scored Nightfall Strikes - Nessus",s:"Nightfall: The Ordeal: Master",t:"29:33"},
    {i:"dungeon",d:"Dungeon - IX Realms",s:"Prophecy",t:"43:22"},
    {i:"explore",d:"Explore - The Moon",s:"",t:"06:14"},
    {i:"storypvecoopheroic",d:"Story - The Cosmodrome",s:"Vendetta",t:"11:50"},
    {i:"ironbanner",d:"Iron Banner - The Crucible",s:"Midtown",t:"03:27"},
    {i:"strikes",d:"Normal Strikes - The Tangled Shore",s:"Broodhold",t:"13:01"},
    {i:"nightmarehunt",d:"Nightmare Hunt - The Moon",s:"Insanity: Legend",t:"08:32"},
    {i:"trialsofosiris",d:"Trials of Osiris - The Crucible",s:"Convergence",t:"04:50"},
    {i:"raid",d:"Raid - Black Garden",s:"Garden of Salvation",t:"1:37:12"}
];
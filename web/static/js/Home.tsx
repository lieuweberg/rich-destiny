import React, { useEffect } from "react";
import { Link } from "react-router-dom";

import PresenceCard, { PresenceCardProps } from "./components/PresenceCard";

import "../css/home.scss";

// @ts-expect-error
import banner from "../images/hero-savathun.webp"

// @ts-expect-error
import home1 from "../images/home1.webp";
// @ts-expect-error
import home2 from "../images/home2.webp";
// @ts-expect-error
import home3 from "../images/home3.webp";

// @ts-expect-error
import homeWaves1 from "../images/home-waves1.svg";

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

    function mouseMove(e: MouseEvent) {
        let hero = document.querySelector("#hero>img");
        let x = (window.innerWidth - e.pageX) / 150;
        let y = (window.innerHeight - e.pageY) / 150;
        // @ts-expect-error
        hero.style.transform = `translateX(${x}px) translateY(${y}px)`;
    }

    useEffect(() => {
        document.addEventListener("mousemove", mouseMove);

        return () => {
            document.removeEventListener("mousemove", mouseMove);
        }
    }, [])
    

    return <>
        <div id="hero">
            <img src={banner} alt="" />
        </div>

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

        <img className="svg-spacer" src={homeWaves1} alt="" />

        <div id="info-cards">
            <div className="card">
                <div className="image">
                    <img width="300px" height="150px" src={home1} alt="screenshot showing tooltip" />
                </div>
                <hr />
                <div>
                    <h3>Forge your fireteam</h3>
                    <p>Anyone playing Destiny 2 on Steam can easily launch the game through your status. <small>(optional)</small></p>
                </div>
            </div>
            {/* <div className="card">
                <div className="image">
                    <img src={home1} alt="screenshot showing join game button"/>
                </div>
                <hr/>
                <div>
                    <h3>Forge your fireteam</h3>
                    <p>No more sharing join codes: have anyone effortlessly join your fireteam by clicking a button in
                        your status. <small>(disabled)</small></p>
                </div>
            </div> */}
            <div className="card">
                <div className="image">
                    <img width="300px" height="150px" src={home2} alt="screenshot showing tooltip"/>
                </div>
                <hr/>
                <div>
                    <h3>Make your class proud</h3>
                    <p>Displays current class and power level as a tooltip on the class icon.</p>
                </div>
            </div>
            <div className="card">
                <div className="image">
                    <img width="300px" height="150px" src={home3} title="credit: tiberaus" alt="image of errors before arrow, checkmarks after"/>
                </div>
                <hr/>
                <div>
                    <h3>It just works</h3>
                    <p>Runs in the background with minimal resource overhead. Auto updates by default. Fully self-contained.</p>
                </div>
            </div>
        </div>

        <img className="svg-spacer flip" src={homeWaves1} alt="" />

        <div className="generic-text top-padding">
            <h1>What's so special about this?</h1>
            <p>After <Link to="/download">downloading</Link>, when first run, this program installs itself into the
            Windows service manager. Every time you now start your computer, this program will
            be started in the background, do some minimal work and then sit idly, waiting until you start Destiny 2.</p>
            <p>All alternatives have a built-in UI (User Interface). This adds a lot of
            overhead while rarely being used. Instead, rich-destiny uses a hosted web-based interface for customising
            settings to keep resource usage, installation size and update size minimal.</p>
        </div>
    </>
}

const examplePresences = [
    {i:"strikes",d:"Normal Strikes - Nessus",s:"The Insight Terminus",t:"08:42"},
    {i:"destinylogo",d:"In Orbit",s:"space 🌌 (customisable)",t:"04:13"},
    {i:"storypvecoopheroic",d:"Story - Europa",s:"The New Kell",t:"13:24"},
    {i:"raid",d:"Last Wish - The Dreaming City",s:"Shuro Chi, The Corrupted (2/5)",t:"02:15:11"},
    {i:"control",d:"Control - The Crucible",s:"Pacifica",t:"05:58"},
    {i:"gambit",d:"Gambit",s:"Emerald Coast",t:"07:26"},
    {i:"crucible",d:"Mayhem - The Crucible",s:"Javelin-4",t:"02:55"},
    {i:"socialall",d:"Social - Earth",s:"Tower",t:"13:17"},
    {i:"raid",d:"Deep Stone Crypt - Europa",s:"Atraks-1, Fallen Exo (2/4)",t:"59:08"},
    {i:"strikes",d:"Nightfall: The Ordeal - The Cosmodrome",s:"Difficulty: Master",t:"29:33"},
    {i:"dungeon",d:"Dungeon - IX Realms",s:"Prophecy",t:"43:22"},
    {i:"explore",d:"Explore - The Moon",s:"",t:"06:14"},
    {i:"lostsector",d:"Lost Sector - Europa",s:"Perdition: Master",t:"11:50"},
    {i:"ironbanner",d:"Iron Banner - The Crucible",s:"Midtown",t:"03:27"},
    {i:"strikes",d:"Normal Strikes - The Tangled Shore",s:"Broodhold",t:"13:01"},
    {i:"nightmarehunt",d:"Nightmare Hunt - The Moon",s:"Insanity: Legend",t:"08:32"},
    {i:"trialsofosiris",d:"Trials of Osiris - The Crucible",s:"Convergence",t:"04:50"},
    {i:"raid",d:"Vault of Glass - Venus",s:"Atheon, Time's Conflux (5/5)",t:"1:37:12"}
];
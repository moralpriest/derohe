# Statement on how DERO proceeds

**Issued by the DERO Foundation — 31 August 2026**

---

## What this document is

The forensic record of the consensus bug closed by HF3 is published in full today, in two documents:

- **[The DERO Double-Spend, Explained](PUBLIC-EXPLAINER.html)** — the mechanism, the arithmetic, and the measurement showing the fix holds.
- **[The Tracked Flow](TRACKED-FLOW.html)** — every transaction we can establish, seed to destination, checkable against any node.

They are not summarised here and they have not been softened. They contain the parts that reflect badly on the project, the parts nobody can see, and the places where our own numbers stop short. Read them first; this document does not repeat them.

In short: at least **23,886,517.3 DERO** of spendable coin was created.

What follows is only the part the reports deliberately left open. Both end at the same place: that coin is real, spendable, permanent unless burned, and no forensic work can settle what the project should now do about it. That is a governance call. This is it.

---

## Decision 1 — DERO moves forward. No rollback.

There will be **no rollback, no relaunch, and no reissue**.

We understand what that costs and we are not going to pretend otherwise. It means coins created out of nothing remain in circulation, and it means whoever created them keeps whatever they still hold. Nobody in this project finds that acceptable in principle. We reached it anyway, and we would rather state the reasoning than have it inferred.

**A rollback does not reach the attacker.** The value realised from these coins left as other assets, off-chain, weeks ago. Rewinding the chain does not claw any of that back. What it does reach is everyone who bought DERO in good faith since — people who did nothing wrong and had no way of knowing. A rollback selects those people to absorb the loss while the attacker keeps the proceeds either way. We are not willing to make that trade, and no version of it we could construct changed that.

**A rollback would claim a certainty we do not have.** Fourteen affected transactions cannot be decoded. We do not know what they moved. We also cannot rule out that other issues exist in a codebase of this age. To rewind the chain and present it as a cure would be to tell holders the problem is solved when we cannot know that. If something further surfaced afterwards, we would have spent the project's remaining credibility on a false assurance — and we would have deserved to lose it. Acting blindly is unavoidable here; acting blindly *and* drastically is not.

**A rollback creates the worse precedent.** Establishing that DERO rewinds its ledger when an outcome is unpopular makes that an expectation rather than an exception. Every future dispute would arrive with it attached, and the next one would carry more weight precisely because it had been done once. It also relocates the chain's security from consensus to whoever can most effectively lobby for a rewind, which is a weaker place to keep it than where it is now.

**Shutdown and relaunch is the same decision with less to show for it.** It carries identical fallout and additionally forfeits the chain's history and continuity.

The fallout from this decision is real and we expect it. So was the fallout from every alternative in front of us. Given a choice between paths that all cost the community something, we took the one that does not also require us to overstate what we know.

---

## Decision 2 — the created supply goes into the code

Not rolling back does not mean moving on quietly. The created coin is real, it is spendable, and it is going to be counted in public, in the software, indefinitely.

The reports set out three available answers on what the project publishes as its supply. We are taking the third.

**`total_supply` stays exactly as it is.** It is the emission schedule, computed from block height, and it is precise at the one job it does. Changing it breaks the only thing it is currently good for and substitutes one misleading number for another. The reports explain at length why it was never evidence in either direction; we are not going to make it into evidence now by editing it.

**A second value will be reported alongside it.** The daemon's `DERO.GetInfo` response will carry an additional, explicitly-labelled figure recording the coin established by forensic analysis as created by the double-application, net of verified burns. It sits beside the schedule figure rather than replacing it, so a caller sees both and neither is quietly conflated with the other. The final field name is settled at implementation; the reports and the field will use the same label.

### The rules that value operates under

Stated now, so they bind us later.

1. **Achievable forensics only.** The figure carries what has been derived and checked against the chain. Nothing estimated, nothing inferred, nothing rounded up to make a point, nothing included because it would look more thorough.
2. **It is a floor, and it is labelled as one.** Fourteen affected transactions cannot be read. Whatever they moved is not in the number and cannot be. The label says so; it does not need a reader to have found the footnote.
3. **Verified burns are netted. Unverified ones are not.** The five million burn is bound by consensus to coins that genuinely left a funded account. That these particular wallets paid it is not proven and cannot be from the chain alone. Both figures are carried, and the difference between them is stated rather than resolved by preference.
4. **It changes only by commit, with the reference data attached.** Every movement in that value ships as a commit carrying the forensic material justifying it. The evidence lives in the repository, in version control, with a history anyone can walk. A number that can be edited without evidence is not a record.
5. **It rolls forward.** If anything further is established — a decoded transaction, a further burn, a recovered amount — the value is updated and new reference data is attached the same way. This is a maintained record, not a one-time announcement. If a comparable event ever occurs again, this is the mechanism that handles it.

### What this value is not

**It is not a balance census.** DERO account balances are encrypted. There is no census in the software and no obvious way to build one on a chain of this design. Nobody can count DERO's balances — not us, not exchanges, not anyone claiming otherwise. This figure is a tracked, evidenced, floor-bounded record of what forensic work has established, which is the honest ceiling on what anyone can offer.

**It is not an audited circulating supply and should not be published as one.** Integrators displaying DERO supply are asked to read both values and their definitions before treating either as one.

We are aware this is the least flattering option available, in that it commits the project to carrying a number documenting its own worst incident, in code, indefinitely. That is the point. If DERO is here for the protocol and the science, then tracking a created-supply figure with its evidence attached is no different in kind from anything else the chain has to account for. There is no honesty caveat we can find in doing it, and several in not doing it.

---

## Decision 3 — the forensic record becomes repository material

The two reports are being committed to the repository, not merely published as a post. They are the reference data behind the tracked value, and they travel with it, so the evidence and the number cannot quietly drift apart. Future updates attach their material the same way.

Anyone wishing to dispute the figure will find everything needed to do so in the same place as the figure itself. That is deliberate.

---

## Operational

1. **Node operators:** run **release 152 or later** if you are not already.
2. **Exchanges and integrators:** both values will be returned by `DERO.GetInfo` once the release ships. `total_supply` is unchanged in value and meaning. Please read Decision 2 before treating either as a supply figure. Technical questions are welcome directly.
3. The forensic reports land in the repository today; the daemon change follows in the next release.

---

## Closing

Everything factual about this incident is in the two reports, checkable against a node today, and we would rather you checked it than took our word for any of it.

Everything in this document is a judgement call. They are ours, we have set out the reasoning so they can be argued with on the merits, and we are not going to pretend they were free of cost. What we will not do is let the number disappear into a footnote. It goes in the code, with its evidence, and it stays there.

To everyone who did the forensic work, ran the controls, and pushed back hard on the comfortable answers: this record exists because of you.

— **The DERO Foundation**

---
---

# Short-form versions

## Matrix / Discord announcement

> **DERO Foundation — how the project proceeds**
>
> The full forensic record of the consensus bug closed by HF3 is published today. Both documents, unedited, including the parts nobody can see and the places our numbers stop short:
> → The DERO Double-Spend, Explained: [LINK]
> → The Tracked Flow: [LINK]
>
> In short: at least 23,886,517.3 DERO of spendable coin was created.
>
> Those reports establish the facts and deliberately stop where forensics stop. This is the part they left open:
>
> **1. No rollback, no relaunch, no reissue.**
> A rollback doesn't reach the attacker — that value left as other assets weeks ago. It reaches everyone who bought in good faith since. It would also claim a certainty we don't have: 14 affected transactions can't be decoded. And it would establish that DERO rewinds its ledger when an outcome is unpopular, making that an expectation rather than an exception for every dispute that follows.
>
> We expect fallout from this. Every alternative carried the same fallout and additionally required us to overstate what we know.
>
> **2. The created supply goes into the code.**
> `total_supply` stays exactly as it is — it's the emission schedule computed from height, and editing it would only make a different number misleading. A second, explicitly-labelled value will be reported alongside it, carrying the forensically-established created coin net of verified burns.
>
> Rules that value runs under: achievable forensics only; labelled as a floor; verified burns netted, unverified ones not; changed **only** by commit with the reference data attached; updated as anything further is established. It is not a balance census — none is possible on an encrypted chain, by us or anyone.
>
> **3. The reports become repository material**, committed as the reference data behind that value, so the evidence and the number can't drift apart.
>
> **Node operators: run release 152 or later.**

## X / Twitter thread

**1/**
The DERO Foundation's full forensic record of the consensus bug closed by HF3 is published today — unedited, including the parts that reflect badly on us.

The reports establish the facts. This thread is the part they left open: what we're doing about it. 🧵

**2/**
Read these first. Everything factual is here, checkable against a node:

→ The DERO Double-Spend, Explained: [LINK]
→ The Tracked Flow, every transaction seed to destination: [LINK]

In short: at least 23,886,517.3 DERO of spendable coin was created. We'd rather you checked that against a node than took our word.

**3/**
**No rollback. No relaunch. No reissue.**

It means created coin stays in circulation. Nobody here finds that acceptable in principle. We reached it anyway, and here's why.

**4/**
A rollback doesn't reach the attacker. That value left as other assets, off-chain, weeks ago.

It reaches the people who bought in good faith since — who did nothing wrong and had no way of knowing.

It picks them to absorb the loss. The attacker profits either way.

**5/**
It would also claim a certainty we don't have.

14 affected transactions can't be decoded. We don't know what they moved.

Rewinding and calling it a cure tells holders it's solved when we can't know that. If something surfaced after, we'd have spent our credibility on a false assurance.

**6/**
And it sets the worse precedent.

Establishing that DERO rewinds its ledger when an outcome is unpopular makes that an expectation, not an exception. Every dispute after this one arrives with it attached — carrying more weight, because it was done once.

**7/**
**Instead, the created supply goes into the code.**

`total_supply` is unchanged. It's the emission schedule computed from block height — the reports explain why it was never evidence in either direction, and editing it now would only relocate the problem.

**8/**
A second, explicitly-labelled value will be reported alongside it: coin established by forensics as created by the double-application, net of verified burns.

Achievable forensics only. Labelled as a floor. Verified burns netted, unverified ones not.

**9/**
It changes **only** by commit, with the forensic reference data attached in-repo. A number that can be edited without evidence isn't a record.

And it rolls forward — updated whenever something further is established. Maintained, not announced once.

**10/**
What it is not: a balance census.

DERO balances are encrypted. Nobody can count them — not us, not exchanges, not anyone claiming otherwise.

It's a tracked, evidenced floor. That's the honest ceiling on what anyone can offer, so it's what we're offering.

**11/**
This commits the project to carrying a number documenting its own worst incident, in code, indefinitely.

That's the point.

**Node operators: run release 152 or later.**

## Exchange / integrator notice

**Subject: DERO — forensic disclosure, project position, and a change to reported supply values**

The full forensic record of the consensus issue closed by HF3 is published today at [LINK] and [LINK]. It includes every recoverable transaction ID and is checkable against any node. It establishes that at least 23,886,517.3 DERO of spendable coin was created. This notice covers only the decisions taken and the integration impact.

**Chain position.** DERO is not rolling back, relaunching or reissuing. The reasoning is set out in full in the accompanying statement.

**Node requirement.** **Nodes must run release 152 or later.**

**Change to reported values.** In an upcoming release, `DERO.GetInfo` will return an additional, explicitly-labelled value alongside `total_supply`, recording coin established by forensic analysis as created by the exploit, net of verified burns.

- `total_supply` is unchanged in both value and meaning. It is the emission schedule computed from block height; it reads no balances and never has. It should not be published as a circulating or auditable supply figure — before this incident or after it.
- The new value is a maintained forensic floor, **not** a balance census. DERO balances are encrypted; no census is possible on this chain by any party. Fourteen affected transactions cannot be decoded, so the figure understates by an unknown amount.
- It changes only via commit with the supporting forensic data committed alongside, and is updated as further material is established.

Integrators displaying DERO supply are asked to review both values and their definitions before publishing either. Technical questions are welcome directly.

## Repository release-note blurb

> **Reported supply values**
>
> `DERO.GetInfo` now returns an additional value alongside `total_supply`, recording coin established by forensic analysis as created by the double-application closed in HF3, net of verified burns.
>
> `total_supply` is unchanged — it remains the emission schedule computed from block height and reads no balances.
>
> The new value is a forensic floor, not a balance census: fourteen affected transactions cannot be decoded and are not represented in it. It is modified only by commit, with the supporting reference data committed alongside, and is updated as further material is established. See `docs/forensics/`.

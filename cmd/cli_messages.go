package cmd

import (
	"math/rand"
)

var scanStartMsgs = []string{
	"Scanning your Mac setup...",
	"Gathering your tools and configs...",
	"Snapshotting your dev world...",
	"Mapping the bones of your setup...",
}

var scanCompleteMsgs = []string{
	"Your setup is looking great! Profile saved.",
	"All done! Your dev setup is safely captured.",
	"Locked and loaded. Your setup is backed up.",
	"Scan complete. Your dev skeleton is safely captured.",
}

var restoreStartMsgs = []string{
	"Hang tight, setting up your Mac...",
	"Time to make this Mac feel like home...",
	"Rebuilding your dev environment...",
	"Fleshing out your Mac...",
}

var restoreCompleteMsgs = []string{
	"All done! Your Mac is feeling like home again.",
	"Welcome back! Everything's set up and ready to go.",
	"Fresh Mac, same great setup. Happy coding!",
	"The bones are back! Restart your terminal to apply changes.",
}

var cloneCompleteMsgs = []string{
	"Profile cloned! Give it a look before restoring.",
	"Got it! Someone's setup is now in your pocket.",
	"Cloned and ready. Trust but verify before restoring.",
	"Nice find! Review it with 'skel show' first.",
}

var publishCompleteMsgs = []string{
	"Your setup is live! Share the link with your team.",
	"Published! Now anyone can clone your setup.",
	"Out in the wild. Your dev setup is officially shareable.",
	"Shipped! Your foundations are ready for the world.",
}

func randomMessage(msgs []string) string {
	return msgs[rand.Intn(len(msgs))]
}

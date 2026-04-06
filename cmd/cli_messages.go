package cmd

import (
	"math/rand"
)

var scanStartMsgs = []string{
	"Scanning your Mac setup...",
	"Gathering your tools and config files...",
	"Capturing a fresh snapshot of your setup...",
	"Mapping your development environment...",
}

var scanCompleteMsgs = []string{
	"Scan complete. Your setup profile is saved.",
	"All set. Your development setup is backed up.",
	"Done. Your environment snapshot is ready.",
	"Great news: your setup is safely captured.",
}

var restoreStartMsgs = []string{
	"Hang tight, setting up your Mac...",
	"Getting this Mac ready for development...",
	"Restoring your development environment...",
	"Applying your saved setup...",
}

var restoreCompleteMsgs = []string{
	"All done. Your Mac is ready to code.",
	"Welcome back. Everything is set up and ready.",
	"Fresh Mac, familiar setup. Happy coding!",
	"Restore complete. Restart your terminal to apply changes.",
}

var cloneCompleteMsgs = []string{
	"Profile cloned. Give it a quick review before restoring.",
	"Done. The profile is ready to inspect and restore.",
	"Clone complete. Verify first, then restore with confidence.",
	"Nice find. Review it with 'skel show' before restoring.",
}

var cloneStartMsgs = []string{
	"Fetching profile data from GitHub Gist...",
	"Bringing this setup into your profile library...",
	"Cloning profile details and validating safety...",
}

var importStartMsgs = []string{
	"Importing profile data from file...",
	"Loading and validating profile contents...",
	"Bringing this setup into your local profiles...",
}

var updateStartMsgs = []string{
	"Refreshing your setup snapshot...",
	"Checking what changed in your environment...",
	"Updating your profile with current system data...",
}

var deleteStartMsgs = []string{
	"Double-checking before deleting this profile...",
	"One quick confirmation, then clean up...",
	"Tidying up your saved profiles...",
}

var exportStartMsgs = []string{
	"Preparing a shareable profile export...",
	"Packaging your setup into a JSON file...",
	"Creating an export you can share with others...",
}

var doctorStartMsgs = []string{
	"Running a quick health check on this profile...",
	"Checking required tools before restore...",
	"Verifying your machine is restore-ready...",
}

var publishCompleteMsgs = []string{
	"Published. Share the link with your team.",
	"Your profile is live and ready to clone.",
	"Done. Your setup is now shareable.",
	"Published successfully. Others can clone it now.",
}

func randomMessage(msgs []string) string {
	return msgs[rand.Intn(len(msgs))]
}

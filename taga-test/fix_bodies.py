import os
import re

dir_path = "/home/sudhan_dev/Downloads/code/nammataga/taga-test/pkg/ui/actions"

replacements = [
    # member_actions.go fixes
    (
        r'r\.WaitForElementAndCapture\(mp\.Page, "css:body", 5 \* time\.Second, "Access_Home"\)',
        r'r.WaitForElementAndCapture(mp.Page, "css:[data-testid=\'home-link\']", 5 * time.Second, "Access_Home")'
    ),
    (
        r'r\.WaitForElementAndCapture\(mp\.Page, "css:body", 5 \* time\.Second, screenshotName\)',
        r'r.WaitForElementAndCapture(mp.Page, "css:[data-testid=\'testid-member-profile-card\'], [data-testid=\'testid-member-subscriptions-button\']", 5 * time.Second, screenshotName)'
    ),
    (
        r'r\.WaitForElementAndCapture\(mp\.Page, "css:body", 5 \* time\.Second, fmt\.Sprintf\("TAGATower_GuestBooking_Complete_%s", roomID\)\)',
        r'r.WaitForElementAndCapture(mp.Page, "css:li[data-sonner-toast]", 5 * time.Second, fmt.Sprintf("TAGATower_GuestBooking_Complete_%s", roomID))'
    ),
    (
        r'r\.WaitForElementAndCapture\(mp\.Page, "css:body", 5 \* time\.Second, fmt\.Sprintf\("TAGATower_AllRoomBooking_Complete_%s", roomID\)\)',
        r'r.WaitForElementAndCapture(mp.Page, "css:li[data-sonner-toast]", 5 * time.Second, fmt.Sprintf("TAGATower_AllRoomBooking_Complete_%s", roomID))'
    ),
    # admin_actions.go fixes
    (
        r'r\.WaitForElementAndCapture\(ap\.Page, "css:body", 5 \* time\.Second, screenshotName\)',
        r'r.WaitForElementAndCapture(ap.Page, "css:[data-testid=\'testid-manage-content-button\']", 5 * time.Second, screenshotName)'
    ),
    # public_actions.go fixes (though failures are fine to just capture body, let's keep body for failures)
]

for f in ["member_actions.go", "admin_actions.go"]:
    p = os.path.join(dir_path, f)
    with open(p, "r") as file:
        content = file.read()
    
    for old_regex, new_str in replacements:
        content = re.sub(old_regex, new_str, content)
        
    with open(p, "w") as file:
        file.write(content)

# Now specifically rewrite the VisitAllMemberPages loop to use a map of locators
member_path = os.path.join(dir_path, "member_actions.go")
with open(member_path, "r") as file:
    content = file.read()

old_loop_def = """	pages := []struct {
		TestID         string
		ScreenshotName string
		WaitTime       time.Duration
	}{
		{"testid-home-button", "Home_Page", 2 * time.Second},
		{"testid-office-bearers-button", "Office_Bearers_Page", 2 * time.Second},
		{"testid-resources-button", "Resources_Page", 2 * time.Second},
		{"testid-events-button", "Events_Gallery_Page", 2 * time.Second},
		{"testid-upcoming-events-button", "Upcoming_Events_Page", 2 * time.Second}, // Sub-tab of Events
		{"testid-taga-towers-button", "TAGATowers_Page", 2 * time.Second},
		{"testid-grievance-button", "Grievance_Page", 2 * time.Second},
		{"testid-membership-button", "Member_Profile_Page", 2 * time.Second},          // Membership default is Profile
		{"testid-member-subscriptions-button", "Subscriptions_Page", 2 * time.Second}, // Sub-tab of Membership
		{"testid-member-announcements-button", "Announcements_Page", 2 * time.Second}, // Sub-tab of Membership
	}

	for _, page := range pages {
		if err := mp.Page.ClickByTestID(page.TestID, mp.DefaultTimeout); err != nil {
			r.Status = "failed"
			r.Error = err
			r.Advice = append(r.Advice, fmt.Sprintf("Advice: Verify '%s' element exists", page.TestID))
			return
		}

		r.WaitForElementAndCapture(mp.Page, "css:[data-testid='testid-member-profile-card'], [data-testid='testid-member-subscriptions-button']", 5 * time.Second, page.ScreenshotName)
	}"""

new_loop_def = """	pages := []struct {
		TestID         string
		ScreenshotName string
		WaitTime       time.Duration
		Locator        string
	}{
		{"testid-home-button", "Home_Page", 2 * time.Second, "css:[data-testid='home-link']"},
		{"testid-office-bearers-button", "Office_Bearers_Page", 2 * time.Second, "css:[data-testid='testid-office-bearers-district-select']"},
		{"testid-resources-button", "Resources_Page", 2 * time.Second, "css:[data-testid='testid-resource-central-button']"},
		{"testid-events-button", "Events_Gallery_Page", 2 * time.Second, "css:[data-testid='testid-events-tabs-list']"},
		{"testid-upcoming-events-button", "Upcoming_Events_Page", 2 * time.Second, "css:[data-testid='testid-upcoming-events-tab-content']"},
		{"testid-taga-towers-button", "TAGATowers_Page", 2 * time.Second, "css:[data-testid='testid-taga-towers-page']"},
		{"testid-grievance-button", "Grievance_Page", 2 * time.Second, "css:body"},
		{"testid-membership-button", "Member_Profile_Page", 2 * time.Second, "css:[data-testid='testid-member-profile-card']"},
		{"testid-member-subscriptions-button", "Subscriptions_Page", 2 * time.Second, "css:[data-testid='testid-subscription-card']"},
		{"testid-member-announcements-button", "Announcements_Page", 2 * time.Second, "css:[data-testid='testid-announcement-card']"},
	}

	for _, page := range pages {
		if err := mp.Page.ClickByTestID(page.TestID, mp.DefaultTimeout); err != nil {
			r.Status = "failed"
			r.Error = err
			r.Advice = append(r.Advice, fmt.Sprintf("Advice: Verify '%s' element exists", page.TestID))
			return
		}

		r.WaitForElementAndCapture(mp.Page, page.Locator, 5 * time.Second, page.ScreenshotName)
	}"""

content = content.replace(old_loop_def, new_loop_def)
with open(member_path, "w") as file:
    file.write(content)

print("Done fixing css:body references.")

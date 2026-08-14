import os
import re

dir_path = "/home/sudhan_dev/Downloads/code/nammataga/taga-test/pkg/ui/actions"

mapping = {
    # Public & Nav
    "GoToHome_Success": "css:[data-testid='home-link']",
    "GoToHome_Failure": "css:body",
    "OfficeBearers_Success": "css:[data-testid='testid-office-bearers-district-select']",
    "OfficeBearers_Failure": "css:body",
    "Events_Success": "css:[data-testid='testid-events-container']",
    "Events_Failure": "css:body",
    
    # Admin Login
    "AdminLogin_ModalOpen": "css:[role='dialog']",
    "AdminLogin_OpenFailure": "css:body",
    "AdminLogin_Success": "css:[data-testid='testid-admin-panel-button']",
    "AdminLogin_Failure": "css:body",
    "AdminLogin_HomePage": "css:[data-testid='testid-admin-panel-button']",
    "AdminLogout_HomePage": "css:[data-testid='testid-admin-login-button']",
    
    # Admin Panel
    "AdminPanel_Loaded": "css:[data-testid='testid-manage-content-button']",
    "AdminPanel_AfterAddMember": "css:[data-testid='testid-manage-content-button']",
    "AdminPanel_AfterBulkUpload": "css:[data-testid='testid-manage-content-button']",
    
    # Admin Modals & Toasts
    "AddMember_ModalOpen": "css:[role='dialog']",
    "AddMember_FormFilled": "css:[data-testid='testid-add-member-name-input']",
    "AddMember_SuccessToast": "css:li[data-sonner-toast]",
    
    "AnnouncementModal_Open": "css:[role='dialog']",
    "BulkUpload_ModalOpen": "css:[role='dialog']",
    "BulkUpload_FileSelected": "css:[data-testid='testid-bulk-upload-submit-button']",
    "BulkUpload_SuccessToast": "css:li[data-sonner-toast]",
    
    "DistrictBearers_ModalOpen": "css:[role='dialog']",
    "EditMember_EditFormOpen": "css:[role='dialog']",
    "EditMember_SaveToast": "css:li[data-sonner-toast]",
    
    "EventCreate_ModalOpen": "css:[role='dialog']",
    "GalleryUpload_ModalOpen": "css:[role='dialog']",
    "MembershipReport_ModalOpen": "css:[role='dialog']",
    "MembershipReport_PeriodSelected": "css:[data-testid='testid-membership-report-download-button']",
    "ResourceUpload_ModalOpen": "css:[role='dialog']",

    "MembershipReport_Downloaded": "css:[role='dialog']",
    "MemberListTable_Downloaded": "css:[data-testid='testid-manage-content-button']",
    
    # Member Login & Access
    "MemberLogin_ModalOpen": "css:[role='dialog']",
    "MemberLogin_Failure": "css:body",
    "MemberLogin_Success": "css:[data-testid='testid-logout-button']",
    "MemberLogin_Dashboard": "css:[data-testid='testid-logout-button']",
    "MemberLogin_Filled_Form": "css:[role='dialog']",
    "MemberLogout_HomePage": "css:[data-testid='testid-member-login-button']",
    
    "Access_Events": "css:h1",
    "Access_Home": "css:body",
    "Access_OfficeBearers": "css:[data-testid='testid-office-bearers-district-select']",
    "Access_Profile": "css:button[role='tab']",
    
    "ChangePassword_Filled": "css:[data-testid='testid-change-password-submit']",
    "ChangePassword_Success": "css:li[data-sonner-toast]",
    
    # Modals & Interactions (Member)
    "Payment_Modal_Opened": "css:[role='dialog']",
    "Subscription_Tab_Opened": "css:[data-testid='testid-member-subscriptions-button']",
    
    # TAGA Towers
    "TAGATower_Dashboard": "css:[data-testid='testid-taga-towers-page']",
    "TAGATower_Booking_Modal": "css:[data-testid='testid-room-booking-modal']",
    "TAGATower_Booking_Complete": "css:li[data-sonner-toast]",
    "TAGATower_Booking_Cancelled": "css:li[data-sonner-toast]",
    "TAGATower_SelfMultibooking_Allowed_Failure": "css:li[data-sonner-toast]",
    "TAGATower_SelfMultibooking_Cancelled_Cleanup": "css:li[data-sonner-toast]",
    "TAGATower_SelfMultibooking_Attempt": "css:[data-testid='testid-room-booking-modal']",
    "TAGATower_GenderBooking": "css:[data-testid='testid-room-booking-modal']",
    "TAGATower_GenderMix_Allowed_Failure": "css:li[data-sonner-toast]",
    "TAGATower_Dormitory": "css:li[data-sonner-toast]",
    "TAGATower_10DaysBooking_FullBug": "css:[data-testid='testid-room-booking-modal']",
    "TAGATower_10DaysBooking_Modal": "css:[data-testid='testid-room-booking-modal']",
    "TAGATower_10DaysBooking_Complete": "css:li[data-sonner-toast]",
    "TAGATower_Overlapping_Allowed_Failure": "css:li[data-sonner-toast]"
}

# Regex to match:
# time.Sleep(...)
# r.CaptureScreenshot(page, "string_or_format")
# Sometimes page is ap.Page, p.Page, mp.Page

pattern = re.compile(r'time\.Sleep\([^)]+\)\n\s*r\.CaptureScreenshot\(([^,]+),\s*(.+?)\)')

def get_locator_for_name(name_expr):
    # evaluate common static string or format
    for key, loc in mapping.items():
        if key in name_expr:
            return loc
            
    if "ModalOpen" in name_expr: return "css:[role='dialog']"
    if "Failure" in name_expr: return "css:body"
    if "Toast" in name_expr: return "css:li[data-sonner-toast]"
    if "Loaded" in name_expr or "Success" in name_expr: return "css:body" # fallback
    if "Attempt" in name_expr: return "css:body"
    
    return "css:body"

for f in os.listdir(dir_path):
    if not f.endswith(".go"): continue
    p = os.path.join(dir_path, f)
    with open(p, "r") as file:
        content = file.read()
    
    def replacer(match):
        page_var = match.group(1)
        name_expr = match.group(2)
        loc = get_locator_for_name(name_expr)
        # return r.WaitForElementAndCapture(page, loc, 5 * time.Second, name)
        return f'r.WaitForElementAndCapture({page_var}, "{loc}", 5 * time.Second, {name_expr})'
        
    new_content = pattern.sub(replacer, content)
    
    with open(p, "w") as file:
        file.write(new_content)

print("Done refactoring states.")

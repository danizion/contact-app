import requests
import pytest
import random
import string

BASE_URL = "http://localhost:80"
'''
The following tests are designed to test these scenarios:
User Creation with Invalid Data
Attempt to create a user using an incomplete payload (only providing a username). The expected outcome is a 400 error due to missing required fields such as email and password.

Valid User Creation
Create a user with all the required fields (user_name, email, and password) provided. The test expects a successful creation indicated by a 201 status code.

Duplicate User Creation
Create a user and then attempt to create another user with the same username and email. The first creation should succeed (201), while the duplicate attempt should fail with a 409 conflict error.

Another Valid User Creation
Create an additional valid user using different credentials to ensure that the system accepts multiple distinct users. A 201 status code is expected.

Login with Wrong Credentials
Attempt to log in using a valid user's email but with an incorrect password. The expected outcome is a 401 unauthorized error.

Login with Correct Credentials
Log in using the valid credentials of a previously created primary user. The test verifies a 200 status code and that a token is returned in the response.

Contact Creation Without Authentication
Attempt to create a contact without providing an Authorization header. The expected result is a 401 error indicating that authentication is required.

Contact Creation with Missing Parameters
Attempt to create a contact while omitting a required parameter (in this case, the "first_name"). The server should respond with a 400 error.

Duplicate Contact Creation
After a valid contact is created (via a fixture), try to create the same contact again using identical details. The expected outcome is a 409 error because the contact already exists.

Update Contact with an Invalid ID
Attempt to update a contact using a non-existent contact ID (e.g., 9999). The expected response is a 404 error indicating the contact is not found.

Update Contact with One Field Change
Create a contact and then update it by changing a single field (the phone number). The test verifies a successful update with a 200 status code and an appropriate success message.

Update Contact with All Fields
Create a new contact and then update all its fields (first_name, last_name, phone_number, and address). The expected outcome is a 200 status code along with a confirmation message that the contact was updated successfully.

Get Contacts for a New User with No Contacts
Create a new user who does not have any contacts and then request the contacts list. The test checks that the returned list is empty (length zero).

Get Contacts for a User with Existing Contacts (No Query)
After creating at least one contact for the primary user, fetch all contacts without specifying any query filters. The test expects a non-empty list indicating that contacts exist.

Pagination of Contacts
Ensure that pagination works as expected by:

Checking that the first page (with a limit of 10) returns no more than 10 contacts.
Verifying that the second page is returned as a list (even if it might be empty). This test creates additional contacts if necessary to meet the pagination criteria.
Delete Contact with Invalid ID
Attempt to delete a contact using an ID that does not exist (e.g., 9999). The expected response is a 404 error.

Delete Contact by a Different User
Try to delete a contact (created by the primary user) using the credentials of a secondary user. The test verifies that access control is enforced by expecting a 404 error.

Successful Contact Deletion
Create a new contact and then delete it using the same (primary) user's credentials. The expected outcome is a 200 status code indicating successful deletion.

Repeated Contact Deletion
Create a contact, delete it successfully, and then attempt to delete the same contact again. The second deletion should fail, returning a 404 error because the contact no longer exists.
'''
def random_string(length=6):
    return ''.join(random.choice(string.ascii_lowercase) for _ in range(length))


# ---------------------------
# User Creation Tests
# ---------------------------
def test_create_user_invalid():
    """Attempt to create a user with an invalid request body."""
    url = f"{BASE_URL}/users"
    # Sending incomplete data: missing required fields
    payload = {"user_name": "invaliduser"}
    response = requests.post(url, json=payload)
    assert response.status_code == 400


def test_create_valid_user():
    """Create a valid user."""
    url = f"{BASE_URL}/users"
    username = "testuser_" + random_string()
    email = f"{username}@example.com"
    payload = {
        "user_name": username,
        "email": email,
        "password": "password1"
    }
    response = requests.post(url, json=payload)
    assert response.status_code == 201


def test_create_user_duplicate():
    """Attempt to create a duplicate user (same username/email)."""
    url = f"{BASE_URL}/users"
    username = "duplicateuser_" + random_string()
    email = f"{username}@example.com"
    payload = {
        "user_name": username,
        "email": email,
        "password": "password1"
    }
    # First creation should succeed.
    response = requests.post(url, json=payload)
    assert response.status_code == 201

    # Second creation with the same data should fail.
    response_dup = requests.post(url, json=payload)
    assert response_dup.status_code == 409

def test_create_another_valid_user():
    """Create another valid user with different credentials."""
    url = f"{BASE_URL}/users"
    username = "anotheruser_" + random_string()
    email = f"{username}@example.com"
    payload = {
        "user_name": username,
        "email": email,
        "password": "password1"
    }
    response = requests.post(url, json=payload)
    assert response.status_code == 201

# ---------------------------
# Fixtures for Authentication
# ---------------------------
@pytest.fixture(scope="module")
def primary_user():
    """
    Creates a primary user and logs in.
    Returns a dict with the token and user details.
    """
    url = f"{BASE_URL}/users"
    username = "primary_" + random_string()
    email = f"{username}@example.com"
    password = "password1"
    payload = {
        "user_name": username,
        "email": email,
        "password": password
    }
    r = requests.post(url, json=payload)
    assert r.status_code == 201

    # Login
    login_url = f"{BASE_URL}/login"
    login_payload = {"email": email, "password": password}
    r2 = requests.post(login_url, json=login_payload)
    assert r2.status_code == 200
    token = r2.json().get("token")
    assert token is not None

    return {"token": token, "user_name": username, "email": email}


@pytest.fixture(scope="module")
def secondary_user():
    """
    Creates a secondary user (with no contacts) and logs in.
    This is used to test access control (e.g. deleting a contact from another user).
    """
    url = f"{BASE_URL}/users"
    username = "secondary_" + random_string()
    email = f"{username}@example.com"
    password = "password1"
    payload = {
        "user_name": username,
        "email": email,
        "password": password
    }
    r = requests.post(url, json=payload)
    assert r.status_code == 201

    login_url = f"{BASE_URL}/login"
    login_payload = {"email": email, "password": password}
    r2 = requests.post(login_url, json=login_payload)
    assert r2.status_code == 200
    token = r2.json().get("token")
    return {"token": token, "user_name": username, "email": email}


# ---------------------------
# Login Tests
# ---------------------------
def test_login_wrong_credentials(primary_user):
    """Attempt login with the wrong password."""
    login_url = f"{BASE_URL}/login"
    login_payload = {"email": primary_user["email"], "password": "wrongpassword"}
    response = requests.post(login_url, json=login_payload)
    assert response.status_code == 401

def test_login_correct(primary_user):
    """Login with correct credentials."""
    login_url = f"{BASE_URL}/login"
    login_payload = {"email": primary_user["email"], "password": "password1"}
    response = requests.post(login_url, json=login_payload)
    assert response.status_code == 200
    token = response.json().get("token")
    assert token is not None


# ---------------------------
# Helper function for Contacts
# ---------------------------
def create_contact(token, first_name, last_name, phone_number, address):
    """
    Creates a contact using the provided details.
    The payload follows the structure:
    {
        "first_name": <first_name>,
        "last_name": <last_name>,
        "phone_number": <phone_number>,
        "address": <address>
    }
    """
    url = f"{BASE_URL}/contacts"
    headers = {"Authorization": f"Bearer {token}"}
    payload = {
        "first_name": first_name,
        "last_name": last_name,
        "phone_number": phone_number,
        "address": address
    }
    return requests.post(url, json=payload, headers=headers)


def update_contact(token, contact_id, data):
    url = f"{BASE_URL}/contacts/{contact_id}"
    headers = {"Authorization": f"Bearer {token}"}
    return requests.patch(url, json=data, headers=headers)


def delete_contact(token, contact_id):
    url = f"{BASE_URL}/contacts/{contact_id}"
    headers = {"Authorization": f"Bearer {token}"}
    return requests.delete(url, headers=headers)


# ---------------------------
# Contact Creation Tests
# ---------------------------
def test_create_contact_no_auth():
    """Try creating a contact without an Authorization header."""
    url = f"{BASE_URL}/contacts"
    payload = {
        "first_name": "contact1",
        "last_name": "bd",
        "phone_number": "1234567890",
        "address": "new address"
    }
    response = requests.post(url, json=payload)  # No auth header provided
    assert response.status_code == 401


def test_create_contact_missing_params(primary_user):
    """
    Attempt to create a contact with missing parameters.
    For example, omit the 'first_name'.
    """
    url = f"{BASE_URL}/contacts"
    headers = {"Authorization": f"Bearer {primary_user['token']}"}
    payload = {
        # "first_name" is missing
        "last_name": "bd",
        "phone_number": "1234567890",
        "address": "new address"
    }
    response = requests.post(url, json=payload, headers=headers)
    assert response.status_code == 400



@pytest.fixture(scope="module")
def contact1(primary_user):
    """Creates a valid contact for the primary user."""
    response = create_contact(primary_user["token"], "contact1", "bd", "1234567890", "new address")
    assert response.status_code == 201
    assert "Contact created successfully" in response.json().get("message", "")
    # Assume the response returns an 'id'; if not, default to 1.
    return response.json().get("id", 1)


def test_create_contact_duplicate(primary_user, contact1):
    """Attempt to create a duplicate contact with the same details."""
    response = create_contact(primary_user["token"], "contact1", "bd", "1234567890", "new address")
    assert response.status_code == 409


@pytest.fixture(scope="module")
def additional_contacts(primary_user):
    """
    Creates additional contacts for the primary user (contact2, contact3, contact4)
    and returns their IDs.
    """
    contact_ids = []
    for name in ["contact2", "contact3", "contact4"]:
        response = create_contact(primary_user["token"], name, "bd", "1234567890", "new address")
        assert response.status_code == 201
        contact_ids.append(response.json().get("contact_id"))
    return contact_ids


# ---------------------------
# Contact Update Tests
# ---------------------------
def test_update_contact_invalid_id(primary_user):
    """Attempt to update a contact using an invalid ID."""
    response = update_contact(primary_user["token"], 9999, {"phone_number": "0987654321"})
    assert response.status_code == 404





def test_update_contact_one_value(primary_user, contact1):
    """Update a contact with one value (e.g. phone_number only)."""
    response = create_contact(primary_user["token"], "name", "bd", "1234567890", "new address")

    response = update_contact(primary_user["token"], response.json().get("contact_id"), {"phone_number": "1112223333"})
    assert response.status_code == 200
    assert "Contact updated successfully" in response.json().get("message", "")


def test_update_contact_all_values(primary_user, contact1):
    """Update a contact with all values (first_name, last_name, phone_number, and address)."""
    name = random_string()
    response = create_contact(primary_user["token"], name, "bd", "1234567890", "new address")
    print(response)
    response = update_contact(
        primary_user["token"],
        response.json().get("contact_id"),
        {
            "first_name": "contact1_updated",
            "last_name": "bd",
            "phone_number": "4445556666",
            "address": "updated address"
        }
    )

    assert response.status_code == 200
    assert "Contact updated successfully" in response.json().get("message", "")


# ---------------------------
# Get Contacts Tests
# ---------------------------
def test_get_contacts_empty():
    """
    Create a new user with no contacts and then fetch contacts.
    Expect an empty list.
    """
    # Create a user with no contacts.
    url = f"{BASE_URL}/users"
    username = "emptyuser_" + random_string()
    email = f"{username}@example.com"
    password = "password1"
    payload = {
        "user_name": username,
        "email": email,
        "password": password
    }
    r = requests.post(url, json=payload)
    assert r.status_code == 201

    # Login
    login_url = f"{BASE_URL}/login"
    login_payload = {"email": email, "password": password}
    r2 = requests.post(login_url, json=login_payload)
    assert r2.status_code == 200
    token = r2.json().get("token")
    headers = {"Authorization": f"Bearer {token}"}

    # Get contacts for this user (should be empty)
    response = requests.get(f"{BASE_URL}/contacts", headers=headers)
    assert response.status_code == 200
    contacts = response.json().get("contacts", [])
    assert isinstance(contacts, list)
    assert len(contacts) == 0


def test_get_contacts_no_query(primary_user):
    """For a user with contacts, get contacts without any query parameters."""
    headers = {"Authorization": f"Bearer {primary_user['token']}"}
    response = create_contact(primary_user['token'], "dvir", "yogev", "123", "tel aviv")
    assert response.status_code == 201
    response = requests.get(f"{BASE_URL}/contacts", headers=headers)
    assert response.status_code == 200
    contacts = response.json().get("items", [])
    assert isinstance(contacts, list)
    # Expect that there is at least one contact.
    assert len(contacts) > 0


def test_pagination_contacts(primary_user):
    """
    Validate pagination by ensuring that when there are more than 10 contacts,
    using 'limit' and 'page' query parameters returns the expected results.
    """
    headers = {"Authorization": f"Bearer {primary_user['token']}"}
    # Get current contacts count
    response = requests.get(f"{BASE_URL}/contacts", headers=headers)
    current_contacts = response.json().get("contacts", [])
    count = len(current_contacts)

    # Create additional contacts if needed to ensure there are more than 10.
    if count < 11:
        for i in range(11 - count):
            create_contact(primary_user["token"], f"pag_contact_{i}", "bd", "0001112222", "new address")

    # Test page 1 with limit 10
    params = {"limit": 10, "page": 1}
    response_page1 = requests.get(f"{BASE_URL}/contacts", headers=headers, params=params)
    assert response_page1.status_code == 200
    contacts_page1 = response_page1.json().get("contacts", [])
    assert len(contacts_page1) <= 10

    # Test page 2 with limit 10
    params = {"limit": 10, "page": 2}
    response_page2 = requests.get(f"{BASE_URL}/contacts", headers=headers, params=params)
    assert response_page2.status_code == 200
    contacts_page2 = response_page2.json().get("contacts", [])
    assert isinstance(contacts_page2, list)


# ---------------------------
# Delete Contact Tests
# ---------------------------
def test_delete_invalid_contact(primary_user):
    """Attempt to delete a contact using an invalid ID."""
    response = delete_contact(primary_user["token"], 9999)
    assert response.status_code == 404


def test_delete_contact_different_user(primary_user, secondary_user, contact1):
    """
    Attempt to delete a contact (contact1) created by the primary user using the secondary user's token.
    This should fail.
    """
    response = delete_contact(secondary_user["token"], contact1)
    assert response.status_code == 404


def test_delete_contact_success(primary_user):
    """Create a contact and then delete it successfully."""
    response = create_contact(primary_user["token"], "delete_contact", "bd", "7778889999", "new address")
    assert response.status_code == 201
    contact_id = response.json().get("contact_id")
    del_response = delete_contact(primary_user["token"], contact_id)
    assert del_response.status_code == 200


def test_delete_same_contact_again(primary_user):
    """
    Delete a contact and then try deleting it again.
    The second deletion should return a 404.
    """
    response = create_contact(primary_user["token"], "delete_again", "bd", "7778880000", "new address")
    assert response.status_code == 201
    contact_id = response.json().get("contact_id")
    del_response = delete_contact(primary_user["token"], contact_id)
    assert del_response.status_code == 200

    # Attempt to delete the same contact a second time
    del_response2 = delete_contact(primary_user["token"], contact_id)
    assert del_response2.status_code == 404

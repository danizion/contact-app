import requests
import pytest
import json
import time
from datetime import datetime

BASE_URL = "http://localhost:8080"

@pytest.fixture(scope="session")
def auth_token():
    """
    Creates a user (if needed) and logs in to obtain a JWT token.
    """
    # Create user
    create_user_payload = {
        "username": "testuser",
        "email": "test@example.com",
        "password": "testpass"
    }
    
    # Create user and check response
    create_response = requests.post(f"{BASE_URL}/users", json=create_user_payload)
    if create_response.status_code not in [201, 400]:  # 400 means user might already exist
        pytest.fail(f"Failed to create user: {create_response.text}")

    # Login to obtain token
    login_payload = {
        "email": "test@example.com",
        "password": "testpass"
    }
    
    # Retry login a few times in case of timing issues
    for _ in range(3):
        response = requests.post(f"{BASE_URL}/login", json=login_payload)
        if response.status_code == 200:
            data = response.json()
            token = data.get("token")
            if token:
                return token
        time.sleep(1)  # Wait a second before retrying
    
    pytest.fail(f"Login failed after retries: {response.text}")

@pytest.fixture
def test_contact_data():
    """Fixture to provide test contact data"""
    timestamp = datetime.now().strftime('%Y%m%d%H%M%S')
    return {
        "first_name": f"John_{timestamp}",
        "last_name": f"Doe_{timestamp}",
        "email": f"test_{timestamp}@example.com",
        "phone": "+1234567890"
    }

def test_get_contacts(auth_token):
    """Test GET /contacts endpoint"""
    headers = {"Authorization": f"Bearer {auth_token}"}
    
    response = requests.get(f"{BASE_URL}/contacts", headers=headers)
    assert response.status_code == 200, f"GET /contacts failed: {response.text}"
    data = response.json()
    
    # Validate response structure
    assert isinstance(data, dict), "Response should be a dictionary"

def test_create_contact(auth_token, test_contact_data):
    """Test POST /contacts endpoint"""
    headers = {"Authorization": f"Bearer {auth_token}"}
    
    response = requests.post(f"{BASE_URL}/contacts", json=test_contact_data, headers=headers)
    assert response.status_code == 201, f"Create contact failed: {response.text}"
    data = response.json()
    
    # Validate response structure
    assert "contact_id" in data, "contact_id not returned"
    assert "message" in data, "Success message not returned"
    
    return data["contact_id"]

def test_update_contact(auth_token, test_contact_data):
    """Test PATCH /contacts endpoint"""
    headers = {"Authorization": f"Bearer {auth_token}"}
    
    # Create a contact first
    contact_id = test_create_contact(auth_token, test_contact_data)
    
    # Update the contact
    update_payload = {
        "contact_id": contact_id,
        "first_name": "Jane",
        "last_name": "Smith",
        "email": f"updated_{datetime.now().strftime('%Y%m%d%H%M%S')}@example.com",
        "phone": "+9876543210"
    }
    
    response = requests.patch(f"{BASE_URL}/contacts", json=update_payload, headers=headers)
    assert response.status_code == 200, f"Update contact failed: {response.text}"
    data = response.json()
    
    # Validate response structure
    assert "message" in data, "Update success message missing"

def test_delete_contact(auth_token, test_contact_data):
    """Test DELETE /contacts endpoint"""
    headers = {"Authorization": f"Bearer {auth_token}"}
    
    # Create a contact first
    contact_id = test_create_contact(auth_token, test_contact_data)
    
    # Delete the contact
    delete_payload = {
        "contact_id": contact_id
    }
    response = requests.delete(f"{BASE_URL}/contacts", json=delete_payload, headers=headers)
    assert response.status_code == 200, f"Delete contact failed: {response.text}"
    data = response.json()
    
    # Validate response structure
    assert "message" in data, "Delete success message missing"

def test_error_cases(auth_token):
    """Test error cases for all endpoints"""
    headers = {"Authorization": f"Bearer {auth_token}"}
    
    # Test creating contact with invalid data
    invalid_payload = {
        "first_name": "",  # Empty required field
        "last_name": "Doe",
        "email": "invalid_email",  # Invalid email format
        "phone": "123"  # Invalid phone format
    }
    response = requests.post(f"{BASE_URL}/contacts", json=invalid_payload, headers=headers)
    assert response.status_code == 400, "Invalid data should return 400"
    
    # Test creating duplicate contact
    duplicate_payload = {
        "first_name": "John",
        "last_name": "Doe",
        "email": "duplicate@example.com",
        "phone": "+1234567890"
    }
    # Create first contact
    requests.post(f"{BASE_URL}/contacts", json=duplicate_payload, headers=headers)
    # Try to create duplicate
    response = requests.post(f"{BASE_URL}/contacts", json=duplicate_payload, headers=headers)
    assert response.status_code == 409, "Duplicate contact should return 409"
    
    # Test unauthenticated access
    response = requests.get(f"{BASE_URL}/contacts")
    assert response.status_code == 401, "Unauthenticated access should return 401"

if __name__ == "__main__":
    pytest.main([__file__, "-v"]) 
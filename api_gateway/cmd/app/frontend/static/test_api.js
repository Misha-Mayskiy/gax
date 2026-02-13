// test_api.js
const API_BASE_URL = 'http://localhost:8080';

async function testAPI() {
    console.log('Testing API endpoints...\n');
    
    // Test 1: Health check
    console.log('1. Testing /health:');
    try {
        const health = await fetch(`${API_BASE_URL}/health`);
        console.log('   Status:', health.status);
        console.log('   Response:', await health.text());
    } catch (e) {
        console.log('   Error:', e.message);
    }
    
    // Test 2: List chats
    console.log('\n2. Testing /chat/list:');
    try {
        const chats = await fetch(`${API_BASE_URL}/chat/list?user_id=test`);
        console.log('   Status:', chats.status);
        const data = await chats.json();
        console.log('   Response format:', typeof data);
        console.log('   Data:', JSON.stringify(data, null, 2));
    } catch (e) {
        console.log('   Error:', e.message);
    }
    
    // Test 3: Create direct chat
    console.log('\n3. Testing /chat/create-direct:');
    try {
        const chat = await fetch(`${API_BASE_URL}/chat/create-direct`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                user_id: 'user-1',
                user_id: 'user-2'
            })
        });
        console.log('   Status:', chat.status);
        console.log('   Response:', await chat.text());
    } catch (e) {
        console.log('   Error:', e.message);
    }
}

testAPI();
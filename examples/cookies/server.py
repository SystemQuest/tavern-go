#!/usr/bin/env python3
"""
Simple Flask server for testing session/cookie functionality
Aligned with tavern-py's cookie example
"""
from flask import Flask, request, jsonify, make_response
import uuid

app = Flask(__name__)

# In-memory session storage
sessions = {}


@app.route('/login', methods=['POST'])
def login():
    """Login endpoint that sets a session cookie"""
    data = request.get_json()
    username = data.get('username')
    password = data.get('password')
    
    if username == 'testuser' and password == 'testpass':
        session_id = str(uuid.uuid4())
        sessions[session_id] = {'username': username}
        
        response = make_response(jsonify({
            'message': 'Login successful',
            'username': username
        }))
        response.set_cookie('session_id', session_id)
        response.set_cookie('user_pref', 'theme_dark')  # Additional cookie
        return response, 200
    else:
        return jsonify({'error': 'Invalid credentials'}), 401


@app.route('/api/protected', methods=['GET'])
def protected():
    """Protected endpoint that requires session cookie"""
    session_id = request.cookies.get('session_id')
    
    if not session_id or session_id not in sessions:
        return jsonify({'error': 'Unauthorized'}), 401
    
    user_data = sessions[session_id]
    return jsonify({
        'message': 'Access granted',
        'user': user_data['username'],
        'data': 'secret information'
    }), 200


@app.route('/logout', methods=['POST'])
def logout():
    """Logout endpoint that clears session"""
    session_id = request.cookies.get('session_id')
    
    if session_id in sessions:
        del sessions[session_id]
    
    response = make_response(jsonify({'message': 'Logged out'}))
    response.set_cookie('session_id', '', expires=0)
    return response, 200


if __name__ == '__main__':
    print("Starting cookie test server on http://localhost:5555")
    print("Endpoints:")
    print("  POST /login - Login with username/password")
    print("  GET  /api/protected - Access protected resource")
    print("  POST /logout - Logout")
    app.run(host='0.0.0.0', port=5555, debug=True)

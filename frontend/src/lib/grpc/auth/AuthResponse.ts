// Original file: proto/auth/auth.proto


export interface AuthResponse {
  'access_token'?: (string);
  'token_type'?: (string);
  'expires_in'?: (number);
  'refresh_token'?: (string);
}

export interface AuthResponse__Output {
  'access_token': (string);
  'token_type': (string);
  'expires_in': (number);
  'refresh_token': (string);
}

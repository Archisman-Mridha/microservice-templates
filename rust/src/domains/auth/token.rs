// MIT License
//
// Copyright (c) 2026 OpenMedia
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

use {
  crate::domains::auth::error::Error,
  anyhow::anyhow,
  chrono::Local,
  jsonwebtoken::{DecodingKey, EncodingKey, Header, Validation, decode, encode},
  serde::{Deserialize, Serialize},
  tracing::error
};

pub struct JWTService {
  issuer:              String,
  audiences:           String,
  token_expires_after: usize,

  encoding_key: EncodingKey,
  decoding_key: DecodingKey
}

impl JWTService {
  pub fn new(issuer: String,
             audiences: String,
             token_expires_after: usize,

             encoding_secret: String)
             -> Self {
    let encoding_key = EncodingKey::from_secret(encoding_secret.as_bytes());
    let decoding_key = DecodingKey::from_secret(encoding_secret.as_bytes());

    Self { issuer,
           audiences,
           token_expires_after,

           encoding_key,
           decoding_key }
  }

  pub fn issue(&self, user_id: i64) -> Result<String, Error> {
    let issued_at = Local::now().timestamp() as usize;
    let expires_at = issued_at + self.token_expires_after;

    let claims = Claims { registered: RegisteredClaims { issuer: self.issuer.to_owned(),
                                                         subject: user_id.to_string(),
                                                         audiences: self.audiences.to_owned(),
                                                         issued_at,
                                                         expires_at },
                          custom:     CustomClaims {} };

    encode(&Header::default(), &claims, &self.encoding_key)
      .map_err(|error| Error::Unexpected(anyhow!("Failed generating JWT : {error}")))
  }

  pub fn verify(&self, jwt: &str) -> Result<Claims, Error> {
    let mut validation = Validation::default();
    validation.set_issuer(&[&self.issuer]);
    validation.set_audience(&[&self.audiences]);

    let token_data = decode::<Claims>(jwt, &self.decoding_key, &validation).map_err(|error| {
                       error!("Failed decoding JWT : {error}");

                       Error::DecodingJWTFailed
                     })?;

    let now = Local::now().timestamp() as usize;

    if token_data.claims.registered.expires_at < now {
      return Err(Error::JWTExpired);
    }

    Ok(token_data.claims)
  }
}

/// JSON web tokens (JWTs) claims are pieces of information asserted about a subject.
#[derive(Serialize, Deserialize)]
pub struct Claims {
  #[serde(flatten)]
  pub registered: RegisteredClaims,

  #[serde(flatten)]
  pub custom: CustomClaims
}

/// Registered claims are standard claims registered with the Internet Assigned Numbers
/// Authority (IANA) and defined by the JWT specification to ensure interoperability with
/// third-party, or external, applications.
#[derive(Serialize, Deserialize)]
pub struct RegisteredClaims {
  /// The "iss" (issuer) claim identifies the principal that issued the JWT. The "iss" value is
  /// a case-sensitive string containing a StringOrURI value.
  #[serde(rename = "iss")]
  pub issuer: String,

  /// The "sub" (subject) claim identifies the principal that is the subject of the JWT. The
  /// claims in a JWT are normally statements about the subject.
  #[serde(rename = "sub")]
  pub subject: String,

  /// The "aud" (audiences) claim identifies the recipients that the JWT is intended for. Each
  /// principal intended to process the JWT MUST identify itself with a value in the audience
  /// claim. If the principal processing the claim does not identify itself with a value in the
  /// "aud" claim when this claim is present, then the JWT MUST be rejected. In the general
  /// case, the "aud" value is an array of case-sensitive strings, each containing a StringOrURI
  /// value. In the special case when the JWT has one audience, the "aud" value MAY be a single
  /// case-sensitive string containing a StringOrURI value.
  #[serde(rename = "aud")]
  pub audiences: String,

  #[serde(rename = "iat")]
  pub issued_at: usize,

  /// The "exp" (expiration time) claim identifies the expiration time on or after which the JWT
  /// MUST NOT be accepted for processing. The processing of the "exp" claim requires that the
  /// current date/time MUST be before the expiration date/time listed in the "exp" claim.
  #[serde(rename = "exp")]
  pub expires_at: usize
}

/**
  Custom claims consist of non-registered public or private claims :

    (1) Public claims : You can create custom claims for public consumption, which might
        contain generic information like name and email. If you create public claims, you must
        either register them or use collision-resistant names through namespacing and take
        reasonable precautions to make sure you are in control of the namespace you use.

    (2) Private claims : You can create private custom claims to share information specific to
        your application. For example, while a public claim might contain generic information
        like name and email, private claims would be more specific, such as employee ID and
        department name.
*/
#[derive(Serialize, Deserialize)]
pub struct CustomClaims {}

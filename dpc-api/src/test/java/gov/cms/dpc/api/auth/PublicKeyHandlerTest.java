package gov.cms.dpc.api.auth;

import gov.cms.dpc.api.exceptions.PublicKeyException;
import org.bouncycastle.asn1.x509.SubjectPublicKeyInfo;
import org.junit.jupiter.api.Test;

import javax.crypto.BadPaddingException;
import javax.crypto.Cipher;
import javax.crypto.IllegalBlockSizeException;
import javax.crypto.NoSuchPaddingException;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.security.*;
import java.security.spec.InvalidKeySpecException;
import java.security.spec.X509EncodedKeySpec;
import java.util.Base64;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;

class PublicKeyHandlerTest {

    @Test
    void testValidKey() throws NoSuchAlgorithmException {
        final String encoded = generatePublicKey();
        final String key = String.format("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----\n", encoded);
        PublicKeyHandler.parsePEMString(key);
    }

    @Test
    void testInvalidKey() throws NoSuchAlgorithmException {
        final String encoded = generatePublicKey();
        final String key = String.format("-----BEGIN RSA PUBLIC KEY-----\n%s\n-----END RSA PUBLIC KEY-----\n", encoded);
        assertThrows(PublicKeyException.class, () -> PublicKeyHandler.parsePEMString(key));
    }

    @Test
    void testEmptyKey() {
        assertThrows(PublicKeyException.class, () -> PublicKeyHandler.parsePEMString(""));
    }

    @Test
    void testGarbageKey() {
        assertThrows(PublicKeyException.class, () -> PublicKeyHandler.parsePEMString("This is NOT a real key"));
    }

    @Test
    void testPrivateKey() throws NoSuchAlgorithmException {
        final KeyPairGenerator kpg = KeyPairGenerator.getInstance("RSA");
        final KeyPair keyPair = kpg.generateKeyPair();
        final String encoded = Base64.getMimeEncoder().encodeToString(keyPair.getPrivate().getEncoded());

        final String key = String.format("-----BEGIN RSA PRIVATE KEY-----\n%s\n-----END RSA PRIVATE KEY-----\n", encoded);
        assertThrows(PublicKeyException.class, () -> PublicKeyHandler.parsePEMString(key));
    }

    @Test
    void testEncryptionRoundTrip() throws NoSuchPaddingException, NoSuchAlgorithmException, InvalidKeyException, BadPaddingException, IllegalBlockSizeException, IOException, InvalidKeySpecException {
        final String plainText = "This should be encrypted";
        final Cipher cipher = Cipher.getInstance("RSA");

        final KeyPairGenerator kpg = KeyPairGenerator.getInstance("RSA");
        final KeyPair keyPair = kpg.generateKeyPair();
        final String encoded = Base64.getMimeEncoder().encodeToString(keyPair.getPublic().getEncoded());
        final String key = String.format("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----\n", encoded);
        final SubjectPublicKeyInfo publicKeyInfo = PublicKeyHandler.parsePEMString(key);

        cipher.init(Cipher.ENCRYPT_MODE, keyPair.getPrivate());
        final byte[] cipherText = cipher.doFinal(plainText.getBytes(StandardCharsets.UTF_8));


        // Reconstruct the public key and try to decrypt
        X509EncodedKeySpec spec = new X509EncodedKeySpec(publicKeyInfo.getEncoded());
        final PublicKey rsa = KeyFactory.getInstance("RSA").generatePublic(spec);
        cipher.init(Cipher.DECRYPT_MODE, rsa);
        final byte[] decrypted = cipher.doFinal(cipherText);
        assertEquals(plainText, new String(decrypted), "Should have matching plain text");
    }

    private String generatePublicKey() throws NoSuchAlgorithmException {
        final KeyPairGenerator kpg = KeyPairGenerator.getInstance("RSA");
        final KeyPair keyPair = kpg.generateKeyPair();

        return Base64.getMimeEncoder().encodeToString(keyPair.getPublic().getEncoded());
    }
}
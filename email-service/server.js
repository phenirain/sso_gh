import express from 'express';
import { Resend } from 'resend';
import dotenv from 'dotenv';

dotenv.config();

const app = express();
const resend = new Resend(process.env.RESEND_API_KEY);

app.use(express.json());

// Endpoint для отправки письма со ссылкой сброса пароля
app.post('/send-reset-email', async (req, res) => {
  try {
    const { to, resetLink, login } = req.body;

    if (!to || !resetLink || !login) {
      return res.status(400).json({
        error: 'Missing required fields: to, resetLink, login'
      });
    }

    console.log(`Sending password reset email to: ${to}`);

    const { data, error } = await resend.emails.send({
      from: process.env.FROM_EMAIL,
      to: to,
      subject: 'Сброс пароля - Cosmetics Shop',
      html: generateResetEmailHTML(login, resetLink)
    });

    if (error) {
      console.error('Resend error:', error);
      return res.status(400).json({ error });
    }

    console.log('Email sent successfully:', data.id);
    res.status(200).json({
      success: true,
      messageId: data.id
    });

  } catch (error) {
    console.error('Server error:', error);
    res.status(500).json({
      error: 'Internal server error',
      message: error.message
    });
  }
});

// Генерация HTML письма (минималистичный черно-белый стиль)
function generateResetEmailHTML(login, resetLink) {
  return `
    <!DOCTYPE html>
    <html>
    <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
    </head>
    <body style="margin: 0; padding: 20px; font-family: monospace; background: #fff; color: #000;">
      <div style="max-width: 600px; margin: 0 auto; border: 2px solid #000;">

        <!-- Header -->
        <div style="background: #000; color: #fff; padding: 20px; text-align: center;">
          <div style="font-size: 24px; font-weight: bold;">
            PASSWORD RESET
          </div>
          <div style="margin-top: 10px; font-size: 14px; letter-spacing: 2px;">
            COSMETICS SHOP
          </div>
        </div>

        <!-- Content -->
        <div style="padding: 30px;">
          <div style="margin-bottom: 20px;">
            <strong>Здравствуйте,</strong>
          </div>

          <div style="margin-bottom: 20px;">
            Получен запрос на сброс пароля для аккаунта: <strong>${login}</strong>
          </div>

          <div style="margin-bottom: 30px;">
            Чтобы установить новый пароль, нажмите на кнопку ниже:
          </div>

          <!-- Button -->
          <div style="text-align: center; margin: 30px 0;">
            <a href="${resetLink}"
               style="display: inline-block;
                      background: #000;
                      color: #fff;
                      padding: 15px 40px;
                      text-decoration: none;
                      border: 2px solid #000;
                      font-weight: bold;
                      letter-spacing: 1px;">
              СБРОСИТЬ ПАРОЛЬ
            </a>
          </div>

          <div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #000; font-size: 12px; color: #333;">
            Если вы не запрашивали сброс пароля, просто проигнорируйте это письмо.
          </div>

          <div style="margin-top: 10px; font-size: 12px; color: #333;">
            Или скопируйте ссылку в браузер:<br>
            <span style="word-break: break-all;">${resetLink}</span>
          </div>
        </div>

        <!-- Footer -->
        <div style="background: #f5f5f5; padding: 15px; text-align: center; font-size: 12px; border-top: 1px solid #000;">
          Cosmetics Shop - Your Beauty Destination
        </div>

      </div>
    </body>
    </html>
  `;
}

// Health check
app.get('/health', (req, res) => {
  res.json({ status: 'ok', service: 'email-service' });
});

// Запуск сервера
const PORT = process.env.PORT || 3001;
app.listen(PORT, () => {
  console.log(`Email service running on port ${PORT}`);
  console.log(`Endpoint: http://localhost:${PORT}/send-reset-email`);
});

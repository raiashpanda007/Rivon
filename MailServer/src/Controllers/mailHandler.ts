import { type Request, type Response, type NextFunction, text } from "express";
import { z as zod } from "zod";
import nodemailer from "nodemailer";
import dotenv from "dotenv";
dotenv.config();


const USER_EMAIL = process.env.USER_EMAIL ?? "your@gmail.com";
const EMAIL_APP_PASSWORD = process.env.EMAIL_APP_PASSWORD ?? "email-password";
const MailHandlerTypes = zod.object({
  email: zod.email(),
  subject: zod.string(),
  body: zod.string(),
})

const transporter = nodemailer.createTransport({
  service: 'gmail',
  auth: {
    user: USER_EMAIL,
    pass: EMAIL_APP_PASSWORD
  }
});

const generateEmailTemplate = (content: string) => {
  return `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Rivon Notification</title>
    <style>
        body { margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #000000; color: #ffffff; }
        .container { max-width: 600px; margin: 0 auto; padding: 40px 20px; }
        .card { background-color: #09090b; border: 1px solid #27272a; border-radius: 16px; overflow: hidden; box-shadow: 0 4px 24px rgba(0, 0, 0, 0.5); }
        .header { background-color: #09090b; padding: 32px; text-align: center; border-bottom: 1px solid #27272a; }
        .logo { color: #f97316; font-size: 32px; font-weight: 800; letter-spacing: -1px; text-decoration: none; display: inline-block; }
        .content { padding: 40px 32px; text-align: center; }
        .message-box { background-color: #18181b; border: 1px solid #27272a; border-radius: 12px; padding: 24px; margin: 24px 0; font-size: 24px; font-weight: 600; color: #f97316; letter-spacing: 1px; word-break: break-all; }
        .text { color: #a1a1aa; font-size: 16px; line-height: 1.6; margin-bottom: 24px; }
        .footer { padding: 32px; text-align: center; color: #52525b; font-size: 12px; border-top: 1px solid #27272a; background-color: #09090b; }
    </style>
</head>
<body>
    <div class="container">
        <div class="card">
            <div class="header">
                <div class="logo">Rivon</div>
            </div>
            <div class="content">
                <div class="text">Hello,</div>
                <div class="text">Here is your verification code or message:</div>
                <div class="message-box">
                    ${content}
                </div>
                <div class="text">If you did not request this, please ignore this email.</div>
            </div>
            <div class="footer">
                <p>&copy; ${new Date().getFullYear()} Rivon. All rights reserved.</p>
                <p>Trade Teams. Bet Smarter.</p>
            </div>
        </div>
    </div>
</body>
</html>
  `;
}

export default function MailHandler(req: Request, res: Response, next: NextFunction) {
  const parsedBody = MailHandlerTypes.safeParse(req.body);
  console.log(req.body);
  if (!parsedBody.success) {
    return res.status(422).json({ Message: "Please provide a valid data", Data: parsedBody.error })
  }
  const { body, subject, email } = parsedBody.data;

  const mailOptions = {
    from: USER_EMAIL,
    to: email,
    subject: subject,
    html: generateEmailTemplate(body)
  };
  transporter.sendMail(mailOptions, function (error, info) {
    if (error) {
      console.log(error);
      return res.status(500).json({
        Message: "Internal Server Error ...",
        Data: error
      })
    } else {
      console.log('Email sent: ' + info.response);
      return res.status(200).json({
        Message: "Mail Sent ... "
      })
    }
  });
}


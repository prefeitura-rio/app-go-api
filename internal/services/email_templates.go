package services

import (
	"fmt"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

// EmailTemplate defines the structure for email content
type EmailTemplate struct {
	Subject string
	Body    string
	IsHTML  bool
}

// GetEnrollmentPendingEmailTemplate returns email template for "Em Análise" status
func GetEnrollmentPendingEmailTemplate(inscricao *models.Inscricao, curso *models.Curso, prefrioDomain string) EmailTemplate {
	subject := fmt.Sprintf("Inscrição recebida! - %s", curso.Titulo)

	meusCoursosURL := fmt.Sprintf("https://%s/servicos/cursos/meus-cursos", prefrioDomain)

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <p>Olá, %s!</p>

    <p>Tudo bem? Recebemos seu interesse na atividade <strong>%s</strong>.</p>

    <p>Neste momento, sua inscrição está sendo avaliada pela equipe responsável do(a) <strong>%s</strong>. Fique atento ao seu e-mail: enviaremos uma atualização assim que o resultado da análise for liberado.</p>

    <p>👉 Se preferir, acompanhe o status clicando em "Meus Cursos" na plataforma Oportunidades Cariocas.</p>

    <p>Boa sorte!</p>

    <p>Equipe do Oportunidades Cariocas</p>

    <p><em>Observação: Este é um e-mail automático. Por favor, não o responda.</em></p>
</body>
</html>`,
		inscricao.Name,
		curso.Titulo,
		curso.InstituicaoNome,
	)

	return EmailTemplate{
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}
}

// GetEnrollmentApprovedEmailTemplate returns email template for "Inscrito" status
func GetEnrollmentApprovedEmailTemplate(inscricao *models.Inscricao, curso *models.Curso, prefrioDomain string) EmailTemplate {
	subject := fmt.Sprintf("Parabéns! Sua inscrição foi confirmada - %s", curso.Titulo)

	cursosURL := fmt.Sprintf("https://%s/servicos/cursos", prefrioDomain)

	// Build location and schedule info
	var locationInfo string
	if curso.IsPresencial() {
		locationInfo = fmt.Sprintf("📍 Endereço: %s", curso.GetPresencialAddress())
	} else {
		locationInfo = "📍 Endereço: online"
	}

	var scheduleInfo string
	if curso.HasSchedule() {
		scheduleInfo = fmt.Sprintf("🕒 Horário de início: %s", curso.GetScheduleDescription())
	}

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <p>Olá, %s!</p>

    <p>Temos uma ótima notícia! 🎉 Você está confirmado(a) na atividade <strong>%s</strong>.</p>

    <p><strong>Local e Horário:</strong></p>
    <p>%s<br>
    %s</p>

    <p>A partir de agora, para informações específicas sobre a atividade, o contato será realizado diretamente pela equipe responsável do(a) <strong>%s</strong>. Fique atento(a) ao seu e-mail e/ou telefone.</p>

    <p>Em caso de dúvidas e sugestões, acesse o site e/ou acompanhe as redes sociais do(a) <strong>%s</strong>.</p>

    <p>Aproveite para conferir outros cursos na nossa plataforma <a href="%s" style="color: #0066cc; text-decoration: none;">Oportunidades Cariocas</a>.</p>

    <p>Até a próxima 👋<br>
    Equipe do Oportunidades Cariocas</p>

    <p><em>Observação: Este é um e-mail automático. Por favor, não o responda.</em></p>
</body>
</html>`,
		inscricao.Name,
		curso.Titulo,
		locationInfo,
		scheduleInfo,
		curso.InstituicaoNome,
		curso.InstituicaoNome,
		cursosURL,
	)

	return EmailTemplate{
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}
}

// GetEnrollmentRejectedEmailTemplate returns email template for "Recusado" status
func GetEnrollmentRejectedEmailTemplate(inscricao *models.Inscricao, curso *models.Curso, prefrioDomain string) EmailTemplate {
	subject := fmt.Sprintf("Informações sobre sua inscrição - %s", curso.Titulo)

	cursosURL := fmt.Sprintf("https://%s/servicos/cursos", prefrioDomain)

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <p>Olá, %s!</p>

    <p>Agradecemos o seu interesse na atividade <strong>%s</strong>.</p>

    <p>Após a análise realizada pela instituição responsável, infelizmente, sua inscrição não foi aprovada desta vez.</p>

    <p>A não aprovação ocorreu por não ter sido cumprido(s) algum(ns) dos requisitos ou critérios definidos para a atividade.</p>

    <p>💡 Mas não desanime! Convidamos você a conhecer outras oportunidades disponíveis em nossa plataforma. Fique de olho e não perca nenhuma das atividades oferecidas pela Prefeitura do Rio, você não vai querer ficar de fora, né? 😉</p>

    <p>👉 Link de acesso: <a href="%s" style="color: #0066cc; text-decoration: none;">Oportunidades Cariocas</a></p>

    <p>Estamos aqui na torcida pelo seu sucesso! Te esperamos por lá! 👋</p>

    <p>Até a próxima 👋<br>
    Equipe do Oportunidades Cariocas</p>

    <p><em>Observação: Este é um e-mail automático. Por favor, não o responda.</em></p>
</body>
</html>`,
		inscricao.Name,
		curso.Titulo,
		cursosURL,
	)

	return EmailTemplate{
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}
}

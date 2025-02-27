package locales

var PtBr = map[string]string{
	"Language":                       "pt-BR",
	"Markets":                        "Mercados",
	"Wallets":                        "Carteiras",
	"Notifications":                  "Notificações",
	"Recent Activity":                "Atividade Recente",
	"Sign Out":                       "Sair",
	"Order History":                  "Histórico de Pedidos",
	"load from file":                 "carregar do arquivo",
	"loaded from file":               "carregado do arquivo",
	"defaults":                       "padrões",
	"Wallet Password":                "Senha da Carteira",
	"w_password_helper":              "Este é a senha que você configurou com o software de sua carteira.",
	"w_password_tooltip":             "Deixar senha vazia caso não haja senha necessária para sua carteira.",
	"App Password":                   "Senha do App",
	"app_password_helper":            "Sua senha do app é sempre necessária quando performando operações sensíveis da carteira.",
	"Add":                            "Adicionar",
	"Unlock":                         "Destrancar",
	"Wallet":                         "Carteira",
	"app_password_reminder":          "Sua senha do app é sempre necessária quando performando operações sensíveis da carteira",
	"DEX Address":                    "Endereço DEX",
	"TLS Certificate":                "Certificado TLS",
	"remove":                         "remover",
	"add a file":                     "adicionar um arquivo",
	"Submit":                         "Enviar",
	"Confirm Registration":           "Confirma Registro",
	"app_pw_reg":                     "Informe sua senha do app para confirmar seu registro na DEX.",
	"reg_confirm_submit":             `Quando vc enviar esse formulário, <span id="feeDisplay"></span> DCR será gasto de sua carteira decred para pagar a taxa de registro.`, // update
	"provided_markets":               "Essa DEX provê os seguintes mercados:",
	"accepted_fee_assets":            "This DEX accepts the following fees:",
	"base_header":                    "Base",
	"quote_header":                   "Quote",
	"lot_size_header":                "Tamanho do Lote",
	"lot_size_headsup":               `Todas as trocas são múltiplas do tamanho do lote.`,
	"Password":                       "Senha",
	"Register":                       "Registrar",
	"Authorize Export":               "Autorizar exportação",
	"export_app_pw_msg":              "Informe a senha para confirmar exportação de conta",
	"Disable Account":                "Desativar Conta",
	"disable_app_pw_msg":             "Informe sua senha para desativar conta",
	"disable_dex_server":             "Este servidor DEX pode ser reativado a qualquer momento no futuro (você não terá que pagar a taxa), adicionando-o novamente.",
	"Authorize Import":               "Autorizar Importação",
	"app_pw_import_msg":              "Informe sua senha do app para confirmar importação da conta",
	"Account File":                   "Arquivo da Conta",
	"Change Application Password":    "Trocar Senha do App",
	"Current Password":               "Senha Atual",
	"New Password":                   "Nova Senha",
	"Confirm New Password":           "Confirmar Nova Senha",
	"Cancel Order":                   "Cancelar pedido",
	"cancel_pw":                      "Informe sua senha para cancelar os pedidos que restam",
	"cancel_no_pw":                   "Enviar ordem de cancelamento para o restante.",
	"cancel_remain":                  "A quantidade restante pode ser alterada antes do pedido de cancelamento ser coincididos.",
	"Log In":                         "Logar",
	"epoch":                          "epoque",
	"price":                          "preço",
	"volume":                         "volume",
	"buys":                           "compras",
	"sells":                          "vendas",
	"Buy Orders":                     "Pedidos de Compras",
	"Quantity":                       "Quantidade",
	"Rate":                           "Câmbio",
	"Epoch":                          "Epoque",
	"Limit Order":                    "Ordem Limite",
	"Market Order":                   "Ordem de Mercado",
	"reg_status_msg":                 `Para poder trocar em <span id="regStatusDex" class="text-break"></span>, o pagamento da taxa de registro é necessário <span id="confReq"></span> confirmações.`,
	"Buy":                            "Comprar",
	"Sell":                           "Vender",
	"Lot Size":                       "Tamanho do Lote",
	"Rate Step":                      "Passo de Câmbio",
	"Max":                            "Máximo",
	"lot":                            "lote",
	"Price":                          "Preço",
	"Lots":                           "Lotes",
	"min trade is about":             "troca mínima é sobre",
	"immediate_explanation":          "Se o pedido não preencher completamente durante o próximo ciclo de encontros, qualquer quantia restante não será reservada ou combinada novamente nos próximos ciclos.", // revisar
	"Immediate or cancel":            "Imediato ou cancelar",
	"Balances":                       "Balanços",
	"outdated_tooltip":               "Balanço pode está desatualizado. Conecte-se a carteira para atualizar.",
	"available":                      "disponível",
	"connect_refresh_tooltip":        "Clique para conectar e atualizar",
	"add_a_base_wallet":              `Adicionar uma carteira<br><span data-unit="base"></span><br>`,
	"add_a_quote_wallet":             `Adicionar uma<br><span data-unit="quote"></span><br>carteira`,
	"locked":                         "trancado",
	"immature":                       "imaturo",
	"Sell Orders":                    "Pedido de venda",
	"Your Orders":                    "Seus Pedidos",
	"Type":                           "Tipo",
	"Side":                           "Lado",
	"Age":                            "Idade",
	"Filled":                         "Preenchido",
	"Settled":                        "Assentado",
	"Status":                         "Status",
	"view order history":             "ver histórico de pedidos",
	"cancel order":                   "cancelar pedido",
	"order details":                  "detalhes do pedido",
	"verify_order":                   `Verificar<span id="vSideHeader"></span> Pedido`,
	"You are submitting an order to": "Você está enviando um pedido para",
	"at a rate of":                   "Na taxa de",
	"for a total of":                 "Por um total de",
	"verify_market":                  "Está é uma ordem de mercado e combinará com o(s) melhor(es) pedidos no livro de ofertas. Baseado no atual valor médio de mercado, você receberá", //revisar
	"auth_order_app_pw":              "Autorizar este pedido com a senha do app.",
	"lots":                           "lotes",
	"provied_markets":                "Essa DEX provê os seguintes mercados:",
	"order_disclaimer": `<span class="red">IMPORTANTE</span>: Trocas levam tempo para serem concluídas, e vc não pode desligar o cliente e software DEX,
		ou o <span data-unit="quote"></span> ou <span data-unit="base"></span> blockchain e/ou software da carteira, até os pedidos serem completamente concluídos.
		A troca pode completar em alguns minutos ou levar até mesmo horas.`, //revisar
	"Order":                       "Ordem",
	"see all orders":              "ver todas as ordens",
	"Exchange":                    "Casa de Câmbio",
	"Market":                      "Mercado",
	"Offering":                    "Oferecendo",
	"Asking":                      "Pedindo",
	"Fees":                        "Taxas",
	"order_fees_tooltip":          "Taxas de transações da blockchain, normalmente coletada por mineradores. Decred DEX não coleta taxas de trocas.",
	"Matches":                     "Combinações",
	"Match ID":                    "ID de Combinação",
	"Time":                        "Tempo",
	"ago":                         "atrás",
	"Cancellation":                "Cancelamento",
	"Order Portion":               "Porção do pedido",
	"you":                         "você",
	"them":                        "Eles",
	"Redemption":                  "Rendenção",
	"Refund":                      "Reembolso",
	"Funding Coins":               "Moedas de Financiamento",
	"Exchanges":                   "Casa de câmbios", //revisar
	"apply":                       "aplicar",
	"Assets":                      "Ativos",
	"Trade":                       "Troca",
	"Set App Password":            "Definir senha de aplicativo",
	"reg_set_app_pw_msg":          "Definir senha de aplicativo. Esta senha protegerá sua conta DEX e chaves e carteiras conectadas.",
	"Password Again":              "Senha Novamente",
	"Add a DEX":                   "Adicionar uma DEX",
	"reg_ssl_needed":              "Parece que não temos um certificado SSL para esta DEX. Adicione o certificado do servidor para podermos continuar.",
	"Dark Mode":                   "Modo Dark",
	"Show pop-up notifications":   "Mostrar notificações de pop-up",
	"Account ID":                  "ID da Conta",
	"Export Account":              "Exportar Conta",
	"simultaneous_servers_msg":    "O cliente da DEX suporta simultâneos números de servidores DEX.",
	"Change App Password":         "Trocar Senha do aplicativo",
	"Build ID":                    "ID da Build",
	"Connect":                     "Conectar",
	"Withdraw":                    "Retirar",
	"Deposit":                     "Depositar",
	"Lock":                        "Trancar",
	"New Deposit Address":         "Novo endereço de depósito",
	"Address":                     "Endereço",
	"Amount":                      "Quantia",
	"Reconfigure":                 "Reconfigurar",
	"pw_change_instructions":      "Trocando a senha abaixo não troca sua senha da sua carteira. Use este formulário para atualizar o cliente DEX depois de ter alterado a senha da carteira pela aplicação da carteira diretamente.",
	"New Wallet Password":         "Nova senha da carteira",
	"pw_change_warn":              "Nota: Trocando para uma carteira diferente enquanto possui trocas ativas ou pedidos nos livros pode causar fundos a serem perdidos.",
	"Show more options":           "Mostrar mais opções",
	"seed_implore_msg":            "Você deve ser cuidadoso. Escreva sua semente e salve uma cópia. Caso você perca acesso a essa maquina ou algum outra problema ocorra, você poderá usar sua semente recupear acesso a sua conta DEX e carteiras regitrada. Algumas contas antigas não podem ser recuperadas, e apesar de novas ou não, é uma boa prática salvar backup das contas de forma separada da semente.",
	"View Application Seed":       "Ver semente da aplicação",
	"Remember my password":        "Lembrar senha",
	"pw_for_seed":                 "Informar sua senha do aplicativo para mostrar sua seed. Tenha certeza que mais ninguém pode ver sua tela.",
	"Asset":                       "Ativo",
	"Balance":                     "Balanço",
	"Actions":                     "Ações",
	"Restoration Seed":            "Restaurar Semente",
	"Restore from seed":           "Restaurar da Semente",
	"Import Account":              "Importar Conta",
	"no_wallet":                   "sem carteira",
	"create_a_x_wallet":           "Criar uma Carteira {{.Info.Name}}",
	"dont_share":                  "Não compartilhe e não perca sua seed.",
	"Show Me":                     "Mostre me",
	"Wallet Settings":             "Configurações da Carteira",
	"add_a_x_wallet":              `Adicionar uma carteira <img data-tmpl="assetLogo" class="asset-logo mx-1"> <span data-tmpl="assetName"></span>`,
	"ready":                       "destrancado",
	"off":                         "desligado",
	"Export Trades":               "Exportar Trocas",
	"change the wallet type":      "trocar tipo de carteira",
	"confirmations":               "confirmações",
	"how_reg":                     "Como você pagará a taxa de registro?",
	"All markets at":              "Todos mercados",
	"pick a different asset":      "Escolher ativo diferente",
	"Create":                      "Criar",
	"Register_loudly":             "Registre!",
	"1 Sync the Blockchain":       "1: Sincronizar a Blockchain",
	"Progress":                    "Progresso",
	"remaining":                   "Faltando",
	"2 Fund the Registration Fee": "2: Financie a Taxa de registro",
	"Registration fee":            "Taxa de registro",
	"Your Deposit Address":        "Seu Endereço de Depósito",
	"add a different server":      "Adicionar um servidor diferente",
	"Add a custom server":         "Adicionar um servidor personalizado",
	"plus tx fees":                "+ tx fees",
	"Export Seed":                 "Exportar Seed",
	"Total":                       "Total",
	"Trading":                     "Trocando",
	"Receiving Approximately":     "Recebendo aproximadamente",
	"Fee Projection":              "Projeção de Taxa",
	"details":                     "detalhes",
	"to":                          "para",
	"Options":                     "Opções",
	"fee_projection_tooltip":      "Se as condições da rede não mudarem antes que seu pedido corresponda, O total de taxas realizadas (como porcentagem da troca) deve estar dentro dessa faixa.",
	"unlock_for_details":          "Desbloqueie suas carteiras para recuperar detalhes do pedido e opções adicionais.",
	"estimate_unavailable":        "Orçamentos e opções de pedidos indisponíveis",
	"Fee Details":                 "Detalhes da taxa",
	"estimate_market_conditions":  "As estimativas de melhor e pior caso são baseadas nas condições atuais da rede e podem mudar quando o pedido corresponder.",
	"Best Case Fees":              "Cenário de melhor taxa",
	"best_case_conditions":        "O melhor cenário para taxa ocorre quando todo pedido é correspondido em uma única combinação.",
	"Swap":                        "Troca",
	"Redeem":                      "Resgatar",
	"Worst Case Fees":             "Pior cenário para taxas",
	"worst_case_conditions":       "O pior caso pode ocorrer se a ordem corresponder em um lote de cada vez ao longo de muitos epoques.",
	"Maximum Possible Swap Fees":  "Taxas Máximas de Troca Possíveis",
	"max_fee_conditions":          "Este é o máximo que você pagaria em taxas em sua troca. As taxas são normalmente avaliadas em uma fração dessa taxa. O máximo não está sujeito a alterações uma vez que seu pedido é feito.",
}
